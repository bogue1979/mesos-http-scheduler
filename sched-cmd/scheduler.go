package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/bogue1979/mesos-sqs-poc/client"
	"github.com/bogue1979/mesos-sqs-poc/mesos/mesos"
	sched "github.com/bogue1979/mesos-sqs-poc/mesos/sched"
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
	events     chan *sched.Event
	doneChan   chan struct{}
	acceptNew  bool
}

// New returns a pointer to new Scheduler
func newSched(master string, fw *mesos.FrameworkInfo, cmd *mesos.CommandInfo) *scheduler {
	return &scheduler{
		client:     client.New(master, "/api/v1/scheduler"),
		framework:  fw,
		command:    cmd,
		cpuPerTask: 0.1,
		memPerTask: 64,
		maxTasks:   5,
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
	log.Println("Mesos-Stream-Id:", s.client.StreamID)

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
	c := time.Tick(time.Duration(*waitTime) * time.Second)
	for now := range c {
		if s.acceptNew != true {
			log.Printf("%s scheduler accept new work", now)
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
			log.Println("Received ", len(offers), " offers ")
			go s.offers(offers)

		case sched.Event_RESCIND:
			log.Println("Received rescind offers")

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
			log.Println("HEARTBEAT")
		}

	}
}

var (
	master = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
	//execPath  = flag.String("executor", "./exec", "Path to test executor")
	mesosUser = flag.String("user", "", "Framework user")
	maxTasks  = flag.Int("maxtasks", 5, "Maximal concurrent tasks")
	cmd       = flag.String("cmd", "echo 'Hello World'", "Command to execute")
	waitTime  = flag.Int("wait", 60, "Wait in seconds before launching new Tasks")
)

func init() {
	flag.Parse()
}

func main() {
	if *mesosUser == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal("Unable to determine user")
		}
		*mesosUser = u.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	fw := &mesos.FrameworkInfo{
		User:     mesosUser,
		Name:     proto.String("Go-HTTP Scheduler"),
		Hostname: proto.String(hostname),
	}
	cmdInfo := &mesos.CommandInfo{
		Shell: proto.Bool(true),
		Value: proto.String(*cmd),
	}

	// health for marathon ;-)
	http.HandleFunc("/", root)
	http.HandleFunc("/health", health)
	go http.ListenAndServe(":8080", nil)

	sched := newSched(*master, fw, cmdInfo)
	sched.maxTasks = *maxTasks
	<-sched.start()
}
