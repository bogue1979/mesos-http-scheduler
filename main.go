package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/user"

	"github.com/bogue1979/mesos-http-scheduler/mesos/mesos"
	"github.com/gogo/protobuf/proto"
)

var (
	master    = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
	mesosUser = flag.String("user", "", "Framework user")
	maxTasks  = flag.Int("maxtasks", 5, "Maximal concurrent tasks")
	cmd       = flag.String("cmd", "echo 'Hello World'", "Command to execute")
	debug     = flag.Bool("debug", false, "Print debug logs")
	waitTime  = flag.Int64("wait", 60, "Wait in seconds before launching new tasks")
	cpu       = flag.Float64("cpu", 0.1, "Cpu Resources for one task")
	mem       = flag.Int("mem", 64, "Memory for one task in MB")
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

	// http health endpoint for marathon ;-)
	http.HandleFunc("/", root)
	http.HandleFunc("/health", health)
	go http.ListenAndServe(":8080", nil)

	sched := newSched(*master, fw, cmdInfo, float64(*mem), *cpu, *waitTime)
	sched.maxTasks = *maxTasks
	<-sched.start()
}
