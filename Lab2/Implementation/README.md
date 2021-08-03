The Raft lab is divided into 3 parts: A, B and C.  
* The implementation of part A can be found [here](PartA)
* The implementation of part B can be found [here](PartB)
* The implementation of part C can be found [here](PartC)

#### Basic ideas of raft
* State Maching change table:
  | Current State\Next State | FOLLOWER | CANDIDATE | LEADER |
  | ---------- | :-----------:  | :-----------: | :-----------: |
  | FOLLOWER   | 1,2     | F     |
  | CANDIDATE  | F     | F     |
  | LEADER     |
  
  - 1: Receiving higher term RequestVotes/AppendEntries RPC from peers.
  - 2: Timeout for not receiving any RequestVotes/AppendEntries RPC from peers.
  - 3: 
