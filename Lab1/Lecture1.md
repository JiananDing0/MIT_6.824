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

#### Understanding of MapReduce
* Basic ideas:  
  - **MapReduce** is a distributed system framework that helps distributing work to all machines in the system.
  - The MapReduce computation can be expressesed as two main part: Map and Reduce.
    * **Map**: Takes an input key/value pair and **produces** a set of intermediate key/value pairs. 
    * The MapReduce library groups together all intermediate values associated with the same intermediate key.
    * **Reduce**: Merges all intermediate values associated with the same intermediate key.  
    ```
    In case that MapReduce is only a framework, or library that is implemented to support computation 
    in distributed systems, users can implement their own versions of Map and Reduce function. Then call 
    other functions in the library to support more efficient calculations.
    ```
  - How does it related to distributed system?  
    Programs written in MapReduce style will be automatically **parallelized** and **executed on a large cluster of commodity machines**. The run-time system takes care of the details of partitioning the input data, scheduling the program’s execution across a set of machines, handling machine failures, and managing the required inter-machine communication.  
    ![Overview of the execution process from user of MapReduce library to work distribution:](image/figure1)
    

* Avoid network communications in case that network throughput is limited.
