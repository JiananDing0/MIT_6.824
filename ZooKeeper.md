## Read note of ZooKeeper
#### 1.Introduction
##### Features
* **Wait-free**: no blocking primitives
* **FIFO client ordering**: enables clients to submit operations asynchronously
* **Linearizable writes**: enable efficient implementation of the service
##### Support an API that allow developers to implement their own primitives
> When designing our coordination service, we moved away from implementing specific primitives on the server side, and instead we opted for exposing an API that enables application developers to implement their own primitives.

#### 2.The ZooKeeper Service
##### 2.1 Service Overview
* *Znode* is similar to tree nodes. Zookeeper has a set of znodes to store data.
  - Zookeeper use standard UNIX notation to refer a znode.
  - Every znode can have a child, every znode must have a parent.
  - Znode has 2 types: regular and ephemeral(暂时的). Ephemeral znodes can be removed manually or automatically after the session create it terminates.
  - Sequential flag: if a znode is created with sequential flag set, then its sequence number must be no smaller than any other nodes that are already created under the same parent.
  - Watch flag: if a operation is created with watch flag set, clients who create the operation will receive timely notifications of changes without requiring polling
##### 2.2 Client API
* All client API has a synchronized version and a unsynchronized version. The unsynchronized version uses a callback.
* All update operation expected a version number, and version number not match will result in failure. If version number is -1, there is no version number check.
##### 2.3 Zookeeper Guarantees
* **Asynchronous Linearizable writes**: all requests that update the state of ZooKeeper are serializable and respect precedence;
* **FIFO client order**: all requests from a given client are executed in the order that they were sent by the client
* Sync API is similar to flush, sync causes a server to apply all pending write requests before processing the read without the overhead of a full write.
##### 2.4 Examples of Primitives
* ZooKeeper’s **ordering guarantees** allow efficient reasoning about system state, and **watches(the watch flag)** allow for efficient waiting.
