# mesos-http-scheduler

1. [What](#what)
2. [Why](#why)
3. [How](#how)
4. [Todo](#todo)

## What

... is a simple [mesos framework scheduler](https://mesos.apache.org/documentation/latest/app-framework-development-guide/) currently using the build in command executor.
The basic intention of this framework is to run one or more tasks of one type as [ETL jobs](https://en.wikipedia.org/wiki/Extract,_transform,_load) on a mesos cluster.
The code is based on [mesos-http](https://github.com/vladimirvivien/mesos-http) and extended.

## Why

Currently we are using [marathon](https://mesosphere.github.io/marathon/) to take care of long running services on our mesos cluster.
Some of these services are implemented as long running jobs but basically doing background tasks.
For example taking some messages out of a [SQS](https://aws.amazon.com/sqs/) and processing these messages.
If the SQS is empty the service will sleep some amount of time and try again.
When the service is in sleeping mode the cpu and memory resources, defined for these tasks, will not be released.
Using this mesos framework as scheduler will free up cluster resources usable for other tasks when the job is done and use another resource offer when needed. ( see [Mesos Architecture](http://mesos.apache.org/documentation/latest/architecture/) )

## How

### Installation

When you have a working GO environment a simple go get is sufficient.
Otherwise you can use docker to build the binary.

```
# with go installed

go get github.com/bogue1979/mesos-http-scheduler

# or use docker
mkdir -p out && docker run -v $(pwd)/out:/go/bin golang sh -c "go get github.com/bogue1979/mesos-http-scheduler && chown ${UID}:${GID} /go/bin/mesos-http-scheduler"

```

Basically point the scheduler to the mesos master and tell him how many parallel tasks you want to start in the cluster.

```
# example
./mesos-http-scheduler -master 192.168.102.2:5050 -mem 128 -cpu 0.2 -cmd 'for i in $(seq 1 $(shuf -i 1-20 -n 1)); do echo $i && sleep 1 ; done' -maxtasks 5 -user root -wait 60
2016/11/14 21:51:17 Subscribed: FrameworkID:  377618e9-37ac-4f5b-931f-601af0d7545d-0001
2016/11/14 21:51:17 Task with ID 1479156677302010578 in state RUNNING
2016/11/14 21:51:17 Task with ID 1479156677302022756 in state RUNNING
2016/11/14 21:51:17 Task with ID 1479156677302026166 in state RUNNING
2016/11/14 21:51:17 Task with ID 1479156677302016999 in state RUNNING
2016/11/14 21:51:17 Task with ID 1479156677302029183 in state RUNNING
2016/11/14 21:51:22 Finished task:  1479156677302010578
2016/11/14 21:51:26 Finished task:  1479156677302026166
2016/11/14 21:51:29 Finished task:  1479156677302016999
2016/11/14 21:51:31 Finished task:  1479156677302022756
2016/11/14 21:51:33 Finished task:  1479156677302029183
2016/11/14 21:52:20 Task with ID 1479156740335434519 in state RUNNING
2016/11/14 21:52:20 Task with ID 1479156740335441201 in state RUNNING
2016/11/14 21:52:20 Task with ID 1479156740335447633 in state RUNNING
2016/11/14 21:52:20 Task with ID 1479156740335450764 in state RUNNING
2016/11/14 21:52:20 Task with ID 1479156740335444457 in state RUNNING
2016/11/14 21:52:21 Finished task:  1479156740335444457
2016/11/14 21:52:22 Finished task:  1479156740335447633
2016/11/14 21:52:29 Finished task:  1479156740335441201
2016/11/14 21:52:33 Finished task:  1479156740335450764
2016/11/14 21:52:38 Finished task:  1479156740335434519
```

This command will start 5 tasks distributed over the mesos cluster. When these tasks are finished it will start new tasks ( up to 5 ) after 70 seconds.


### command line parameters

```

Usage of ./mesos-http-scheduler:
  -cmd string
    	Command to execute (default "echo 'Hello World'")
  -cpu float
    	Cpu Resources for one task (default 0.1)
  -debug
    	Print debug logs
  -master string
    	Master address <ip:port> (default "127.0.0.1:5050")
  -maxtasks int
    	Maximal concurrent tasks (default 5)
  -mem int
    	Memory for one task in MB (default 64)
  -user string
    	Framework user
  -wait int
    	Wait in seconds before launching new tasks (default 60)

```

## TODO

1. using docker executor
2. connect to current mesos master when redirected
3. reconnect after leader change
