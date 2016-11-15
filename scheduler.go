package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bogue1979/mesos-http-scheduler/client"
	"github.com/bogue1979/mesos-http-scheduler/mesos/mesos"
	sched "github.com/bogue1979/mesos-http-scheduler/mesos/sched"
	"github.com/gogo/protobuf/proto"
)

// Scheduler represents a Mesos scheduler
type scheduler struct {
	framework    *mesos.FrameworkInfo
	executor     *mesos.ExecutorInfo
	command      *mesos.CommandInfo
	taskLaunched int
	maxTasks     int

	client     *client.Client
	callClient *client.Client
	cpuPerTask float64
	memPerTask float64
	waitTime   int64
	events     chan *sched.Event
	doneChan   chan struct{}
	acceptNew  bool
}

// New returns a pointer to new Scheduler
func newSched(master string, fw *mesos.FrameworkInfo, cmd *mesos.CommandInfo, mem, cpu float64, wait int64) *scheduler {
	return &scheduler{
		client:     client.New(master, "/api/v1/scheduler"),
		framework:  fw,
		command:    cmd,
		cpuPerTask: cpu,
		memPerTask: mem,
		maxTasks:   5,
		waitTime:   wait,
		events:     make(chan *sched.Event),
		doneChan:   make(chan struct{}),
		acceptNew:  true,
	}
}

// start starts the scheduler and subscribes to event stream
// returns a channel to wait for completion.
func (s *scheduler) start() <-chan struct{} {
	if err := s.subscribe(); err != nil {
		log.Fatal(err)
	}
	go s.handleEvents()
	go s.acceptOffers()
	return s.doneChan
}

func (s *scheduler) stop() {
	close(s.events)
}

func (s *scheduler) send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.client.Send(payload)
}

// Subscribe subscribes the scheduler to the Mesos cluster.
// It keeps the http connection opens with the Master to stream
// subsequent events.
func (s *scheduler) subscribe() error {
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.framework,
		},
	}

	resp, err := s.send(call)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Subscribe with unexpected response status: %d", resp.StatusCode)
	}
	debugLog(fmt.Sprintln("Mesos-Stream-Id:", s.client.StreamID))

	go s.qEvents(resp)

	return nil
}

func (s *scheduler) qEvents(resp *http.Response) {
	defer func() {
		resp.Body.Close()
		close(s.events)
	}()
	dec := json.NewDecoder(resp.Body)
	for {
		event := new(sched.Event)
		if err := dec.Decode(event); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}
		s.events <- event
	}
}

func (s *scheduler) acceptOffers() {
	c := time.Tick(time.Duration(s.waitTime) * time.Second)
	for now := range c {
		if s.acceptNew != true {
			debugLog(fmt.Sprintf("%s scheduler accept new work", now))
			s.acceptNew = true
		}
	}
}

func (s *scheduler) handleEvents() {
	defer close(s.doneChan)
	for ev := range s.events {
		switch ev.GetType() {

		case sched.Event_SUBSCRIBED:
			sub := ev.GetSubscribed()
			s.framework.Id = sub.FrameworkId
			log.Println("Subscribed: FrameworkID: ", sub.FrameworkId.GetValue())

		case sched.Event_OFFERS:
			offers := ev.GetOffers().GetOffers()
			debugLog(fmt.Sprintln("Received ", len(offers), " offers "))
			go s.offers(offers)

		case sched.Event_RESCIND:
			debugLog(fmt.Sprintln("Received rescind offers"))

		case sched.Event_UPDATE:
			status := ev.GetUpdate().GetStatus()
			go s.status(status)

		case sched.Event_MESSAGE:
			log.Println("Received message event")

		case sched.Event_FAILURE:
			log.Println("Received failure event")
			fail := ev.GetFailure()
			if fail.ExecutorId != nil {
				log.Println(
					"Executor ", fail.ExecutorId.GetValue(), " terminated ",
					" with status ", fail.GetStatus(),
					" on agent ", fail.GetAgentId().GetValue(),
				)
			} else {
				if fail.GetAgentId() != nil {
					log.Println("Agent ", fail.GetAgentId().GetValue(), " failed ")
				}
			}

		case sched.Event_ERROR:
			err := ev.GetError().GetMessage()
			log.Println(err)

		case sched.Event_HEARTBEAT:
			debugLog(fmt.Sprintln("HEARTBEAT"))
		}
	}
}
