package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/bogue1979/mesos-http-scheduler/mesos/mesos"
	"github.com/gogo/protobuf/proto"
)

var (
	master      = flag.String("master", "127.0.0.1:5050", "Master addresses <ip:port>[,<ip:port>..]")
	mesosUser   = flag.String("user", "", "Framework user")
	maxTasks    = flag.Int("maxtasks", 5, "Maximal concurrent tasks")
	cmd         = flag.String("cmd", "echo 'Hello World'", "Command to execute")
	dockerImage = flag.String("img", "", "Docker image to use ")
	debug       = flag.Bool("debug", false, "Print debug logs")
	waitTime    = flag.Int64("wait", 60, "Wait in seconds before launching new tasks")
	cpu         = flag.Float64("cpu", 0.1, "Cpu Resources for one task")
	mem         = flag.Int("mem", 64, "Memory for one task in MB")
)

func init() {
	flag.Parse()
}

func findMesosMaster(masters string) (string, error) {
	masterservers := strings.Split(masters, ",")

	for _, master := range masterservers {

		resp, err := http.Get("http://" + master + "/api/v1/scheduler")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 405 {
			fmt.Println("Current Master is ", master)
			return master, nil
		}
	}
	return "", fmt.Errorf("could not find master")
}

func main() {

	if *mesosUser == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal("Unable to determine user")
		}
		*mesosUser = u.Username
	}

	if *dockerImage == "" {
		fmt.Println("need docker image name")
		os.Exit(1)
	}

	mmaster, err := findMesosMaster(*master)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

	sched := newSched(mmaster, fw, cmdInfo, float64(*mem), *cpu, *waitTime)
	sched.maxTasks = *maxTasks
	<-sched.start()
}
