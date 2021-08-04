## Read note of ZooKeeper
#### Introduction
##### Features
* **Wait-free**: no blocking primitives
* **FIFO client ordering**: enables clients to submit operations asynchronously
* **Linearizable writes**: enable efficient implementation of the service
##### Support an API that allow developers to implement their own primitives
> When designing our coordination service, we moved away from implementing specific primitives on the server side, and instead we opted for exposing an API that enables application developers to implement their own primitives.
