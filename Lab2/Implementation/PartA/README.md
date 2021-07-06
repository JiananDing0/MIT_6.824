In lab 2A, the only file we needs to implement is ```../raft/raft.go```. Thus, I only upload this file. Clear comments are provided with the solution.

#### List of const/struct/function added
* Const: FOLLOWER, CANDIDATE, LEADER
* Struct: AppendEntriesArgs
* Struct: AppendEntriesReply
* Struct: Log
* Function: candidateElection(chan int)
* Function: callAppendEntries(int, int, int, int) bool
* Function: callRequestVote(int, int, int, int) bool
* Function: execute()
* Function: HandleAppendEntries(AppendEntriesArgs, AppendEntriesReply)
* Function: heartbeatExecute()
* Function: logMoreUpToDate(int, int) bool
* Function: sendAppendEntries(int, AppendEntriesArgs, AppendEntriesReply) bool

#### List of struct/function modified
* Struct: Raft
* Struct: RequestVoteArgs
* Struct: RequestVoteReply
* Function: GetState() (int, bool)
* Function: HandleRequestVote(RequestVoteArgs, RequestVoteReply) --(Rename from function)-> RequestVote
* Function: Make(\[\]\*labrpc.ClientEnd, int, \*Persister, chan ApplyMsg) \*Raft
* Function: sendRequestVote(int, RequestVoteArgs, RequestVoteReply) bool

