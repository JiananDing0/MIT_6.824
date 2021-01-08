package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

// Job ...
// Self implemented data structure, the basic unit for each job
// ID:			unique id for each job
// Type: 		a binary integer variable, 0 for map, 1 for reduce
// Worker:		the number of worker this job is assigned to, -1 for no worker
// Status:		an integer variable, 0 for incomplete (unassigned or error), 1 for assigned, 2 for complete
// Source: 		a string variable store the name of the source file
type Job struct {
	ID     int
	Type   int
	Status int
	Source []string
}

// ApplyArgs ...
// The rpc workers send for applying a job
type ApplyArgs struct{}

// DoneArgs ...
// The rpc workers send for notifying a job is done
// JobID:			the pointer to the corresponding job
// TempFiles: 		the temporary file names
// TargetFiles:		the target output file names
type DoneArgs struct {
	JobID       int
	TempFiles   []string
	TargetFiles []string
}

// Reply ...
// Reply rpc for workers apply for jobs
// Job:			A pointer to the corresponding job
// NReduce: 	Number of reduce jobs, only useful to map workers
type Reply struct {
	NReduce int
	Job     *Job
	result  bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
