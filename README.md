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


### Running framework

Basically point the scheduler to the mesos master and tell him how many parallel tasks you want to start in the cluster.

```
./mesos-http-scheduler -master 10.4.1.10:5050,10.4.2.10:5050,10.4.1.57:5050 -cmd 'for i in $(seq 1 $(shuf -i 1-20 -n 1)); do echo $i &&sleep 1; done' -maxtasks 2 -user root -wait 60 -img meteogroup/centos:7
Current Master is  10.4.1.57:5050
2016/11/24 17:48:55 Subscribed: FrameworkID:  0921d6b7-ff26-46ef-b168-43809b11e76e-0014
2016/11/24 17:48:58 Task with ID 1480006135827894577 in state RUNNING
2016/11/24 17:48:58 Task with ID 1480006135827870836 in state RUNNING
2016/11/24 17:48:59 Finished task:  1480006135827894577
2016/11/24 17:49:16 Finished task:  1480006135827870836
2016/11/24 17:50:11 Task with ID 1480006195824173035 in state RUNNING
2016/11/24 17:50:11 Task with ID 1480006195824190594 in state RUNNING
2016/11/24 17:50:13 Finished task:  1480006195824173035
2016/11/24 17:50:25 Finished task:  1480006195824190594
2016/11/24 17:50:58 Task with ID 1480006255930854880 in state RUNNING
2016/11/24 17:50:59 Task with ID 1480006255930878553 in state RUNNING
2016/11/24 17:51:04 Finished task:  1480006255930878553
2016/11/24 17:51:17 Finished task:  1480006255930854880
```

This command will start 2 tasks distributed over the mesos cluster. When these tasks are finished it will start new tasks ( up to 2 ) after 60 seconds.


### Command line parameters

```

Usage of ./mesos-http-scheduler:
  -cmd string
    	Command to execute (default "echo 'Hello World'")
  -cpu float
    	Cpu Resources for one task (default 0.1)
  -debug
    	Print debug logs
  -img string
    	Docker image to use
  -master string
    	Master addresses <ip:port>[,<ip:port>..] (default "127.0.0.1:5050")
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

1. reconnect after leader change
