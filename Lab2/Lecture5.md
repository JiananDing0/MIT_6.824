# Lecture 5: Go, Threads, and Raft

#### Use of locks
* Locks are meant to protect invariants(properties holds on shared data that multiple have access to). Don't abuse locks that might break the invariants. For example:
```
// A transfer go routine that transfer money between alice and bob. We want the total amount of alice and bob does not change

// The go routine below does not work, because it breaks the atomicity between the increment and decrement.
// Users might acquire the lock once the decrement from Alice complete. But at this moment, increment does not happens yet.
go func() {
  for i := 0; i < 1000; i++ {
    mu.Lock()
    alice -= 1
    mu.Unlock()
    mu.Lock()
    bob += 1
    mu.Unlock()
  }
}
// The go routine below works, because the atomicity between increment and decrement holds.
go func() {
  for i := 0; i < 1000; i++ {
    mu.Lock()
    alice -= 1
    bob += 1
    mu.Unlock()
  }
}
```
* Sometimes, buffered channel might cause problems. As a result, either choose unbuffered channel or using locks. Locks can always substitute channels.
