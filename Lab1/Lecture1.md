# Lecture 1

#### Reason of distributed systems:
* Parallelism between computers to get large job done.
* Higher fault tolenrance for those happens on some computers of the system.
* Isolated computers provide better security.

#### Main challenges:
* Concurrency problems among computers.
* Solutions for partial failures.
* Methods to get higher performance.

#### Terms:
* Scalability: the performance/throughput of the system can be imporved by adding more computers to it.
* Availability: the system is always available unless certain amount of failures happen.
* Recoverability: recover failures happen in the system. Such as restore the storage data by using non-volatile storage or replications.
* MapReduce: a distributed system framework that helps distributing work to all computers in the system.

#### Understanding of MapReduce
* Basic steps: **MapReduce** is a programming model that:
  - Processes a key/value pair to **generate** a set of intermediate key/value pairs. 
  - Use reduce function that **merges** all intermediate values associated with the same intermediate key.
* It is a method related to distributed system published by Google in 2004.
* Avoid network communications in case that network throughput is limited.
