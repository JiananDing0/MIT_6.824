package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

/*
 * Raft is a Go object implementing a single Raft peer.
 */
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Persistent state on all servers:
	currentTerm int
	votedFor    int
	log         []Log
	// Volatile state on all servers:
	commitIndex int // TODO: whether commitIndex and lastAppliedIndex has the same meaning
	lastApplied int
	// Volatile state on leaders:
	nextIndex  []int
	matchIndex []int
	// Extra variables required for data structure
	status           int
	electionTimeout  int
	heartbeatTimeout int
	changeStatusChan chan int
}

/*
 * Const below are enumarations of representations for current client status
 */
const (
	FOLLOWER  = iota
	CANDIDATE = iota
	LEADER    = iota
)

/*
 * RequestVoteArgs is a data structure that stores arguments for RequestVote RPC
 */
type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

/*
 * RequestVoteReply is a data structure that stores results for RequestVote RPC
 */
type RequestVoteReply struct {
	Term        int // Only useful when vote is not granted
	VoteGranted bool
}

/*
 * AppendEntriesArgs is a data structure that stores arguments for AppendEntries RPC
 */
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

/*
 * AppendEntriesReply is a data structure that stores results for AppendEntries RPC
 */
type AppendEntriesReply struct {
	Term    int // Only useful when unsuccess
	Success bool
}

/*
 * Log is the data structure that holds a single entry element stored in servers
 */
type Log struct {
	Command interface{}
	Term    int
}

// GetState return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool

	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isleader = rf.status == LEADER
	return term, isleader
}

// Persist save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// ReadPersist restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// Start...
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

// Execute holds the main execution logic of switching between follower, candidate
// and leader
func (rf *Raft) execute() {
	// For loop iterate term-by-term
	var currentStatus = FOLLOWER
	for !rf.killed() {
		rf.mu.Lock()
		log.Printf("[Server %d]{Term %d} now in status %s\n", rf.me, rf.currentTerm, [...]string{"FOLLOWER", "CANDIDATE", "LEADER"}[currentStatus])
		rf.status = currentStatus
		rf.mu.Unlock()
		if currentStatus == FOLLOWER {

			select {
			case <-rf.changeStatusChan:
				currentStatus = FOLLOWER
			case <-time.After(time.Duration(rf.electionTimeout) * time.Millisecond):
				currentStatus = CANDIDATE
			}

		} else if currentStatus == CANDIDATE {

			requestVoteChan := make(chan int)
			go rf.candidateElection(requestVoteChan)
			select {
			case <-rf.changeStatusChan:
				currentStatus = FOLLOWER
			case <-requestVoteChan:
				currentStatus = LEADER
			case <-time.After(time.Duration(rf.electionTimeout) * time.Millisecond):
				currentStatus = CANDIDATE
			}

		} else if currentStatus == LEADER {

			go rf.heartbeatExecute()
			select {
			case <-rf.changeStatusChan:
				currentStatus = FOLLOWER
			case <-time.After(time.Duration(rf.heartbeatTimeout) * time.Millisecond):
				currentStatus = LEADER
			}

		}
	}
}

// CandidateElection attemps to start a leader election and tally election results
// from peers
func (rf *Raft) candidateElection(ch chan int) {
	var voteLock sync.Mutex
	var voteCond = sync.NewCond(&voteLock)
	var voteCount = 0
	var totalComplete = 0

	rf.mu.Lock()
	rf.currentTerm++
	rf.votedFor = rf.me

	term := rf.currentTerm
	lastIndex := len(rf.log) - 1
	lastTerm := rf.log[lastIndex].Term
	rf.mu.Unlock()

	// Send replies
	for i := 0; i < len(rf.peers); i++ {
		go func(x int) {
			res := rf.callRequestVote(x, term, lastIndex, lastTerm)
			voteLock.Lock()
			defer voteLock.Unlock()
			if res {
				voteCount++
			}
			totalComplete++
			voteCond.Broadcast()
		}(i)
	}
	// Determine vote result, pass value to ch only if it wins the vote
	voteLock.Lock()
	for voteCount*2 <= len(rf.peers) && totalComplete < len(rf.peers) {
		voteCond.Wait()
	}
	if voteCount*2 > len(rf.peers) {
		// log.Printf("[Server %d]{Term %d} election win.\n", rf.me, term)
		ch <- LEADER
	}
	voteLock.Unlock()
	return
}

// CallRequestVote constructs RPC arguments and call sender of RequestVote RPC
func (rf *Raft) callRequestVote(server, term, lastIndex, lastTerm int) bool {
	args := RequestVoteArgs{
		Term:         term,
		CandidateId:  rf.me,
		LastLogIndex: lastIndex,
		LastLogTerm:  lastTerm,
	}
	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(server, &args, &reply)
	if !ok || !reply.VoteGranted {
		return false
	}
	return true
}

// SendRequestVote send requestVote RPC sender to specific server
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.HandleRequestVote", args, reply)
	return ok
}

// HandleRequestVote contains full logic of how to deal with RequestVote RPC from
// peers, capital letter 'H' because it is a public function
func (rf *Raft) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// First determine vote granted result
	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
	} else if args.Term == rf.currentTerm {
		if rf.logMoreUpToDate(args.LastLogIndex, args.LastLogTerm) && (rf.votedFor == -1 || rf.votedFor == args.CandidateId) {
			reply.VoteGranted = true
		} else {
			reply.VoteGranted = false
		}
	} else {
		reply.VoteGranted = true
	}

	// Then, update reply, server data and server status
	if !reply.VoteGranted {
		reply.Term = rf.currentTerm
	} else {
		if rf.status != FOLLOWER && args.Term > rf.currentTerm {
			// fmt.Printf("Server %d trigger case 1\n", rf.me)
			rf.changeStatusChan <- FOLLOWER
		} else if rf.status == FOLLOWER {
			// fmt.Printf("Server %d trigger case 2\n", rf.me)
			rf.changeStatusChan <- FOLLOWER
		}
		rf.currentTerm = args.Term
		rf.votedFor = args.CandidateId
	}
}

// LogMoreUpToDate determines whether the input arguments represent a more up-to-date
// log in comparison to itself. Return true if the log from RPC is at least as
// up-to-date as itself's, otherwise false
func (rf *Raft) logMoreUpToDate(index, term int) bool {
	// No need for locks because it is guaranteed that this function will be called under protection
	lastIndex := len(rf.log) - 1
	lastTerm := rf.log[lastIndex].Term
	if lastTerm < term {
		return true
	}
	if lastTerm == term && lastIndex <= index {
		return true
	}
	return false
}

// HeartbeatExecute activate a non-blocking process that send heartbeat message to peers
func (rf *Raft) heartbeatExecute() {
	rf.mu.Lock()
	term := rf.currentTerm
	prevIndex := len(rf.log) - 1
	prevTerm := rf.log[prevIndex].Term
	rf.mu.Unlock()

	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		go func(x int) {
			rf.callAppendEntries(x, term, prevIndex, prevTerm)
		}(i)
	}
}

// CallAppendEntries constructs arguments and call sender of AppendEntries RPC
func (rf *Raft) callAppendEntries(server, term, prevIndex, prevTerm int) bool {
	args := AppendEntriesArgs{
		Term:         term,
		LeaderId:     rf.me,
		PrevLogIndex: prevIndex,
		PrevLogTerm:  prevTerm,
		Entries:      nil,
		LeaderCommit: 0,
	}
	reply := AppendEntriesReply{}
	ok := rf.sendAppendEntries(server, &args, &reply)
	if !ok || !reply.Success {
		return false
	}
	return true
}

// SendAppendEntries send AppendEntries RPC sender to specific server
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.HandleAppendEntries", args, reply)
	return ok
}

// HandleAppendEntries contains full logic of how to deal with AppendEntries RPC from
// peers, capital letter 'H' because it is a public function
func (rf *Raft) HandleAppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		// Construct reply
		reply.Success = false
		reply.Term = rf.currentTerm
	} else {
		// Construct reply
		reply.Success = true
		// Update server data
		if rf.status == CANDIDATE {
			rf.changeStatusChan <- FOLLOWER
		} else if rf.status == LEADER && args.Term > rf.currentTerm {
			rf.changeStatusChan <- FOLLOWER
		} else if rf.status == FOLLOWER {
			rf.changeStatusChan <- FOLLOWER
		}
		rf.currentTerm = args.Term
	}
	return
}

// Kill...
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// Make...
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	// Initialize variables
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.log = append(rf.log, Log{Command: nil, Term: rf.currentTerm})
	rf.electionTimeout = rand.Intn(300) + 150
	rf.heartbeatTimeout = 100
	rf.status = FOLLOWER
	rf.changeStatusChan = make(chan int)
	go rf.execute()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
