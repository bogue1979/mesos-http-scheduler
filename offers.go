package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	mesos "github.com/bogue1979/mesos-http-scheduler/mesos/mesos"
	sched "github.com/bogue1979/mesos-http-scheduler/mesos/sched"
	"github.com/gogo/protobuf/proto"
)

// Offers handle incoming offers
func (s *scheduler) offers(offers []*mesos.Offer) {
	for _, offer := range offers {
		debugLog(fmt.Sprintln("Processing offer ", offer.Id.GetValue()))

		cpus, mems := s.offeredResources(offer)
		var tasks []*mesos.TaskInfo
		debugLog(fmt.Sprintln("cpus available for tasks: ", cpus, " mems available: ", mems))
		call := &sched.Call{}
		if s.acceptNew == false {
			call = &sched.Call{
				FrameworkId: s.framework.GetId(),
				Type:        sched.Call_DECLINE.Enum(),
				Decline: &sched.Call_Decline{
					OfferIds: []*mesos.OfferID{
						offer.GetId(),
					},
					//Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
				},
			}
		} else {
			for s.taskLaunched < s.maxTasks &&
				cpus >= s.cpuPerTask &&
				mems >= s.memPerTask {

				taskID := fmt.Sprintf("%d", time.Now().UnixNano())
				debugLog(fmt.Sprintln("Preparing task with id ", taskID, " for launch"))
				task := &mesos.TaskInfo{
					Name: proto.String(fmt.Sprintf("task-%s", taskID)),
					TaskId: &mesos.TaskID{
						Value: proto.String(taskID),
					},
					AgentId: offer.AgentId,
					Resources: []*mesos.Resource{
						&mesos.Resource{
							Name:   proto.String("cpus"),
							Type:   mesos.Value_SCALAR.Enum(),
							Scalar: &mesos.Value_Scalar{Value: proto.Float64(s.cpuPerTask)},
						},
						&mesos.Resource{
							Name:   proto.String("mem"),
							Type:   mesos.Value_SCALAR.Enum(),
							Scalar: &mesos.Value_Scalar{Value: proto.Float64(s.memPerTask)},
						},
					},
					Command: s.command,
				}
				tasks = append(tasks, task)
				s.taskLaunched++
				cpus -= s.cpuPerTask
				mems -= s.memPerTask

				if s.taskLaunched == s.maxTasks {
					s.acceptNew = false
				}
			}

			// setup launch call
			call = &sched.Call{
				FrameworkId: s.framework.GetId(),
				Type:        sched.Call_ACCEPT.Enum(),
				Accept: &sched.Call_Accept{
					OfferIds: []*mesos.OfferID{
						offer.GetId(),
					},
					Operations: []*mesos.Offer_Operation{
						&mesos.Offer_Operation{
							Type: mesos.Offer_Operation_LAUNCH.Enum(),
							Launch: &mesos.Offer_Operation_Launch{
								TaskInfos: tasks,
							},
						},
					},
					//Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
				},
			}
		}

		// send call
		resp, err := s.send(call)
		if err != nil {
			log.Println("Unable to send Accept Call: ", err)
			continue
		}
		if resp.StatusCode != http.StatusAccepted {
			log.Printf("Accept Call returned unexpected status: %d", resp.StatusCode)
		}
	}
}

// offeredResources
func (s *scheduler) offeredResources(offer *mesos.Offer) (cpus, mems float64) {
	for _, res := range offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mems += *res.GetScalar().Value
		}
	}
	return
}
