package mr

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Master ...
// Main Master data structure keep track of all jobs
// nMap:			total number of map jobs
// nReduce: 		total number of reduce jobs
// MapQueue:		unassigned map jobs
// ReduceQueue: 	unassigned reduce jobs
// Completequeue: 	completed jobs
type Master struct {
	MapQueue    []Job
	ReduceQueue []Job
	JobLock     *sync.Cond
}

// JobApplyHandler ...
// an RPC handler to handle ApplyArgs type rpc call.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) JobApplyHandler(args *ApplyArgs, reply *Reply) error {
	// Lock here, release the lock after function returns
	m.JobLock.L.Lock()
	defer m.JobLock.L.Unlock()

	// First, find a map job
	for !m.CheckMap() {
		for i := 0; i < len(m.MapQueue); i++ {
			if m.MapQueue[i].Status == 0 {
				m.MapQueue[i].Status = 1
				reply.Job = &m.MapQueue[i]
				reply.NReduce = len(m.ReduceQueue)
				go m.AutoComplete(&m.MapQueue[i])
				return nil
			}
		}
		m.JobLock.Wait()
	}

	// // If not find, wait until all Map job finishes, then assign reduce job
	for !m.CheckReduce() {
		for i := 0; i < len(m.ReduceQueue); i++ {
			if m.ReduceQueue[i].Status == 0 {
				m.ReduceQueue[i].Status = 1
				reply.Job = &m.ReduceQueue[i]
				go m.AutoComplete(&m.ReduceQueue[i])
				return nil
			}
		}
		m.JobLock.Wait()
	}

	// If reach here, then that means no available jobs, return error
	return errors.New("No available jobs")
}

// JobDoneHandler ...
// an RPC handler to handle DoneArgs type rpc call.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) JobDoneHandler(args *DoneArgs, reply *Reply) error {
	// If the arg is to notify master the job is done, then boardcast the job is done
	m.JobLock.L.Lock()
	if args.JobID >= len(m.MapQueue) {
		m.ReduceQueue[args.JobID-len(m.MapQueue)].Status = 2
	} else {
		m.MapQueue[args.JobID].Status = 2
		for i, filename := range args.TargetFiles {
			m.ReduceQueue[i].Source[args.JobID] = filename
		}
	}

	m.JobLock.Broadcast()
	m.JobLock.L.Unlock()
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// Done ...
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	return m.CheckMap() && m.CheckReduce()
}

// CheckMap ...
// Function check whether all map jobs are done
func (m *Master) CheckMap() bool {
	for _, job := range m.MapQueue {
		if job.Status != 2 {
			return false
		}
	}
	return true
}

// CheckReduce ...
// Function check whether all map jobs are done
func (m *Master) CheckReduce() bool {
	for _, job := range m.ReduceQueue {
		if job.Status != 2 {
			return false
		}
	}
	return true
}

// AutoComplete ...
// Force quit a job that takes longer than 10 seconds
func (m *Master) AutoComplete(job *Job) {
	time.Sleep(time.Second * 10)
	if job.Status == 1 {
		m.JobLock.L.Lock()
		job.Status = 0
		m.JobLock.Broadcast()
		m.JobLock.L.Unlock()
	}
}

// MakeMaster ...
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// First, initialize all parameters for master
	m.MapQueue = make([]Job, len(files))
	m.ReduceQueue = make([]Job, nReduce)
	m.JobLock = sync.NewCond(&sync.Mutex{})
	for i, filename := range files {
		m.MapQueue[i] = Job{ID: i, Type: 0, Status: 0, Source: []string{filename}}
	}
	for i := 0; i < nReduce; i++ {
		m.ReduceQueue[i] = Job{ID: i + len(files), Type: 1, Status: 0, Source: make([]string, len(files))}
	}

	// Start listening
	m.server()
	// The master, as an RPC server, will be concurrent; don't forget to lock shared data.

	return &m
}
