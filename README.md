# sched-only

1. [What](#what)
2. [Why](#why)
3. [How](#how)
4. [Todo](#todo)

## What

... is a simple [mesos framework scheduler](https://mesos.apache.org/documentation/latest/app-framework-development-guide/) currently using the build in command executor.
The basic intention of this framework is to run one or more tasks of one type as [ETL jobs](https://en.wikipedia.org/wiki/Extract,_transform,_load) on a mesos cluster.


## Why

Currently we are using [marathon](https://mesosphere.github.io/marathon/) to take care of long running services on our mesos cluster.
Some of these services are implemented as long running jobs but basically doing background tasks.
For example taking some messages out of a [SQS](https://aws.amazon.com/sqs/) and processing these messages.
If the SQS is empty the service will sleep some amount of time and try again.
When the service is in sleeping mode the cpu and memory resources, defined for these tasks, will not be released.
Using this mesos framework as scheduler will free up cluster resources usable for other tasks when the job is done and use another resource offer when needed. ( see [Mesos Architecture](http://mesos.apache.org/documentation/latest/architecture/) )

## How

Basically point the scheduler to the mesos master and tell him how many parallel tasks you want to start in the cluster.

```
# example
./sched-only -master 10.4.1.57:5050 -cmd 'docker run -it alpine echo hello world' -maxtasks 5 -user root -wait 70
```

This command will start 5 tasks distributed over the mesos cluster. When these tasks are finished it will start new tasks ( up to 5 ) after 70 seconds.


### command line parameters

```
Usage of ./sched-only:
  -cmd string
    	Command to execute (default "echo 'Hello World'")
  -master string
    	Master address <ip:port> (default "127.0.0.1:5050")
  -maxtasks int
    	Maximal concurrent tasks (default 5)
  -user string
    	Framework user
  -wait int
    	Wait in seconds before launching new Tasks (default 60)

```

## TODO

1. using docker executor
