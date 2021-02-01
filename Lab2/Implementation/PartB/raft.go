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
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

// ApplyMsg ...
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

// Raft ...
// A Go object implementing a single Raft peer.
// mu: 			Lock to protect shared access to this peer's state
// peers: 		RPC end points of all peers
// persister: 	Object to hold this peer's persisted state
// me: 			This peer's index into peers[]
// dead: 		Set by Kill()
//
// state: 		Current state of the server: 0 for follower, 1 for candidates, 2 for leader
// toFollower:
// toLeader:
// applyCh:		Push committed command index to the state machine for applying
//
// currentTerm:	latest term server has seen (initialized to 0 on first boot, increases monotonically)
// votedFor: 	candidateId that received vote in current term (or null if none)
// log:			log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)
// commitIndex: index of highest log entry known to be committed (initialized to 0, increases monotonically)
// lastApplied: index of highest log entry applied to state machine (initialized to 0, increases monotonically)
//
// nextIndex:	for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
// matchIndex:	for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)
//
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int
	dead      int32

	// Modifications for 2A, 2B
	state      int
	toFollower chan int
	toLeader   chan int
	applyCh    chan ApplyMsg

	currentTerm int
	votedFor    int
	logs        []LogEntry
	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int
}

// GetState ...
// return currentTerm and whether this server
// believes it is the leader.
//
func (rf *Raft) GetState() (int, bool) {
	// Modifications for 2A
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term := rf.currentTerm
	isleader := (rf.state == 2)

	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
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

//
// restore previously persisted state.
//
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

// RequestVoteArgs ...
// Term: 			candidate’s term
// CandidateID: 	candidate requesting vote
// LastLogIndex: 	index of candidate’s last log entry (§5.4)
// LastLogTerm: 	term of candidate’s last log entry (§5.4)
//
type RequestVoteArgs struct {
	// Modifications for 2A, 2B
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

// RequestVoteReply ...
// Term: 			currentTerm, for candidate to update itself
// VoteGranted: 	true means candidate received vote
//
type RequestVoteReply struct {
	// Modifications for 2A
	Term        int
	VoteGranted bool
}

// AppendEntriesArgs ...
// Term:			Leader’s term
// LeaderID:		Support follower can redirect clients
// PrevLogIndex:	Index of log entry immediately preceding new ones
// PrevLogTerm:	 	Term of prevLogIndex entry
// Entries[]: 		Log entries to store (empty for heartbeat; may send more than one for efficiency)
// LeaderCommit: 	Leader’s commitIndex
//
type AppendEntriesArgs struct {
	// Modifications for 2A, 2B
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesReply ...
// Term: 	Current term, for leader to update itself
// Success: True if follower contained entry matching prevLogIndex and prevLogTerm
//
type AppendEntriesReply struct {
	// Modifications for 2A, 2B
	Term    int
	Success bool
}

// LogEntry ...
// Command: Command for state machine
// Term: 	Term when entry was received by leader
// Index:	Index of the log entry, which the first index is 1
//
type LogEntry struct {
	// Modifications for 2A, 2B
	Command interface{}
	Term    int
	Index   int
}

// RequestVote ...
// RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Debug code
	fmt.Printf("Server %d in term %d receive vote request from server %d, term %d, lastLogIndex %d, lastLogTerm %d\n", rf.me, rf.currentTerm, args.CandidateID, args.Term, args.LastLogIndex, args.LastLogTerm)
	// fmt.Println("Lock at RequestVote handler required")
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Ignore request with term less than or equal to its current term
	if args.Term <= rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	// Refuse request that not contains more up-to-date logs
	reply.Term = args.Term
	if (args.LastLogTerm < rf.logs[rf.commitIndex].Term) || (args.LastLogTerm == rf.logs[rf.commitIndex].Term && args.LastLogIndex < rf.commitIndex) {
		reply.VoteGranted = false
		return
	}

	// Otherwise, vote for the candidate
	reply.VoteGranted = true

	rf.currentTerm = args.Term
	rf.votedFor = args.CandidateID
	rf.toFollower <- 0
}

// sendRequestVote ...
// Code to send a RequestVote RPC to a server.
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// AppendEntries ...
// AppendEntries RPC handler.
//
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// fmt.Printf("Server %d in term %d, log %v receive arg in term %d, prevLogIndex %d, prevLogTerm %d\nand entries %v\n", rf.me, rf.currentTerm, rf.logs, args.Term, args.PrevLogIndex, args.PrevLogTerm, args.Entries)

	// Ignore request with term less than its current term
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	rf.currentTerm = args.Term
	reply.Term = args.Term
	rf.toFollower <- 0

	fmt.Printf("Server %d passes case 0\n", rf.me)

	// Return directly on heart beats, commit verified indexes
	if args.Entries == nil {
		for ; rf.commitIndex < args.LeaderCommit && rf.commitIndex < len(rf.logs)-1; rf.commitIndex++ {
			rf.applyCh <- ApplyMsg{
				Command:      rf.logs[rf.commitIndex+1].Command,
				CommandIndex: rf.commitIndex + 1,
				CommandValid: true,
			}
		}
		reply.Success = false
		return
	}

	fmt.Printf("Server %d passes case 1\n", rf.me)

	// Check prevLogIndex and prevLogTerm to see whether they match
	if (args.PrevLogIndex > rf.commitIndex) || (args.PrevLogIndex == rf.commitIndex && args.PrevLogTerm != rf.logs[args.PrevLogIndex].Term) {
		reply.Success = false
		return
	}

	fmt.Printf("Server %d passes case 2\n", rf.me)

	if (args.PrevLogIndex < rf.commitIndex) || (len(rf.logs) > args.PrevLogIndex+len(args.Entries)) {
		reply.Success = true
		return
	}

	fmt.Printf("Server %d passes case 3\n", rf.me)

	// If match, remove invalid, append new, then return, commit all these commands through applyCh
	rf.logs = rf.logs[:args.PrevLogIndex+1]
	rf.logs = append(rf.logs, args.Entries...)
	reply.Success = true
	fmt.Printf("Log on server %d: %v\n", rf.me, rf.logs)
	return
}

// sendAppendEntries ...
// Code to send a AppendEntries RPC to a server.
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// Start ...
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
	// Modifications for 2B.
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state != 2 {
		return -1, -1, false
	}

	// fmt.Printf("Now server %d receives command %v\n", rf.me, command)
	nextEntry := LogEntry{Command: command, Term: rf.currentTerm, Index: len(rf.logs)}
	rf.logs = append(rf.logs, nextEntry)

	// commitCount := 0
	// commitBool := false
	// commitLock := sync.Mutex{}

	for i := 0; i < len(rf.peers); i++ {
		appendEntriesArgs := AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderID:     rf.me,
			LeaderCommit: rf.commitIndex,
			PrevLogIndex: rf.commitIndex,
			PrevLogTerm:  rf.logs[rf.commitIndex].Term,
			Entries:      rf.logs[rf.commitIndex+1:],
		}
		if i != rf.me {
			go func(x int, rf *Raft) {
				for !rf.killed() {
					appendEntriesReply := AppendEntriesReply{}
					if rf.sendAppendEntries(x, &appendEntriesArgs, &appendEntriesReply) {
						if appendEntriesReply.Success {
							if appendEntriesArgs.Entries[len(appendEntriesArgs.Entries)-1].Index > rf.matchIndex[x] {
								rf.mu.Lock()
								rf.matchIndex[x] = appendEntriesArgs.Entries[len(appendEntriesArgs.Entries)-1].Index
								rf.nextIndex[x] = rf.matchIndex[x] + 1
								for {
									rf.commitIndex++
									commitCount := 1
									for i := 0; i < len(rf.peers); i++ {
										if rf.me != i && rf.matchIndex[i] >= rf.commitIndex {
											commitCount++
										}
									}
									if commitCount*2 > len(rf.peers) {
										rf.applyCh <- ApplyMsg{
											Command:      rf.logs[rf.commitIndex].Command,
											CommandIndex: rf.commitIndex,
											CommandValid: true,
										}
									} else {
										rf.commitIndex--
										break
									}
								}
								rf.mu.Unlock()
							}
							break
						}
						if !appendEntriesReply.Success && appendEntriesReply.Term == appendEntriesArgs.Term {
							rf.mu.Lock()
							rf.nextIndex[x]--
							appendEntriesArgs.LeaderCommit = rf.commitIndex
							appendEntriesArgs.PrevLogIndex = rf.nextIndex[x] - 1
							appendEntriesArgs.PrevLogTerm = rf.logs[rf.nextIndex[x]-1].Term
							appendEntriesArgs.Entries = rf.logs[rf.nextIndex[x]:]
							rf.mu.Unlock()
						} else {
							break
						}
					}
					// time.Sleep(10 * time.Millisecond)
				}
			}(i, rf)
		}
	}
	return len(rf.logs) - 1, rf.currentTerm, true
}

// Kill ...
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

// nextState ...
// Return the next state of the server.
//
func (rf *Raft) nextState(timeOut int) {
	var res int
	select {
	case <-rf.toFollower:
		res = 0
	case <-rf.toLeader:
		// fmt.Printf("Server %d becomes leader\n", rf.me)
		res = 2
	case <-time.After(time.Duration(timeOut) * time.Millisecond):
		if rf.state == 0 || rf.state == 1 {
			res = 1
		} else if rf.state == 2 {
			res = 2
		}
	}
	rf.mu.Lock()
	rf.state = res
	rf.mu.Unlock()
}

// Make ...
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

	// Modifications for 2A, 2B
	rf.state = 0
	rf.toFollower = make(chan int)
	rf.toLeader = make(chan int)

	rf.currentTerm = 0
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.votedFor = -1
	rf.logs = make([]LogEntry, 0)
	// Append a initial command with index 0 and term 0
	rf.logs = append(rf.logs, LogEntry{
		Command: nil,
		Index:   0,
		Term:    0,
	})

	// Initialize applyCh
	rf.applyCh = applyCh

	go func(raft *Raft) {
		// Set election timeout frequency
		lowRange := 250
		highRange := 400
		elecTimeOut := rand.Intn(highRange-lowRange) + lowRange

		// Set heartbeat frequency
		heartBeatFreq := 100

		// Start the server execution, the loop will execute forever until killed
		// The server has 3 different state, we needs to do different stuff
		for !raft.killed() {
			raft.mu.Lock()
			if raft.state == 0 {
				raft.mu.Unlock()
				timeOut := elecTimeOut
				// Wait and change to next state, the function is guaranteed to return
				raft.nextState(timeOut)
			} else if raft.state == 1 {
				// Increment term and vote for himself
				raft.currentTerm++
				raft.votedFor = me
				requestVoteArgs := RequestVoteArgs{
					Term:         raft.currentTerm,
					CandidateID:  raft.me,
					LastLogIndex: len(raft.logs) - 1,
					LastLogTerm:  raft.logs[len(raft.logs)-1].Term,
				}
				raft.mu.Unlock()

				voteCount := 1
				voteBool := false
				voteLock := sync.Mutex{}
				// Send vote request to all other servers
				for i := 0; i < len(raft.peers); i++ {
					if i != raft.me {
						go func(x int, vCount *int, vLock *sync.Mutex, vBool *bool, raft *Raft) {
							requestVoteReply := RequestVoteReply{}
							if raft.sendRequestVote(x, &requestVoteArgs, &requestVoteReply) {
								if requestVoteReply.VoteGranted {
									vLock.Lock()
									*vCount++
									if *vCount*2 > len(raft.peers) && !*vBool {
										*vBool = true
										// Reinitialize matchedIndex and nextIndex
										raft.mu.Lock()
										raft.matchIndex = make([]int, len(raft.peers))
										raft.nextIndex = make([]int, len(raft.peers))
										for i := 0; i < len(raft.peers); i++ {
											raft.matchIndex[i] = 0
											raft.nextIndex[i] = raft.commitIndex + 1
										}
										raft.toLeader <- 0
										raft.mu.Unlock()
									}
									vLock.Unlock()
								}
								if requestVoteReply.Term > requestVoteArgs.Term {
									raft.mu.Lock()
									raft.currentTerm = requestVoteReply.Term
									raft.toFollower <- 0
									raft.mu.Unlock()
								}
							}
						}(i, &voteCount, &voteLock, &voteBool, raft)
					}
				}
				timeOut := elecTimeOut
				// Wait and change to next state, the function is guaranteed to return
				raft.nextState(timeOut)
			} else if raft.state == 2 {
				// Send heart beat to all other servers
				appendEntriesArgs := AppendEntriesArgs{
					Term:         raft.currentTerm,
					LeaderID:     raft.me,
					PrevLogIndex: -1,
					PrevLogTerm:  -1,
					Entries:      nil,
					LeaderCommit: raft.commitIndex,
				}
				raft.mu.Unlock()

				for i := 0; i < len(raft.peers); i++ {
					if i != raft.me {
						go func(x int, raft *Raft) {
							appendEntriesReply := AppendEntriesReply{}
							if raft.sendAppendEntries(x, &appendEntriesArgs, &appendEntriesReply) {
								if appendEntriesReply.Term > appendEntriesArgs.Term {
									raft.mu.Lock()
									raft.currentTerm = appendEntriesReply.Term
									raft.toFollower <- 0
									raft.mu.Unlock()
								}
							}
						}(i, raft)
					}
				}
				timeOut := heartBeatFreq
				// Wait and change to next state, the function is guaranteed to return
				raft.nextState(timeOut)
			}
		}
	}(rf)

	// Your initialization code here (2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
