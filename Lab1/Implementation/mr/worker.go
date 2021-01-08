package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
)

// KeyValue ...
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// ByKey ...
// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// Worker ...
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
	RPCCall(mapf, reducef)
}

// RPCCall ...
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func RPCCall(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// declare an argument structure.
	getJobArg := ApplyArgs{}

	// declare a reply structure.
	reply := Reply{}

	// send the RPC request, wait for the reply.
	for call("Master.JobApplyHandler", &getJobArg, &reply) {
		doneJobArg := DoneArgs{JobID: reply.Job.ID}
		// Do job
		if reply.Job.Type == 0 {
			// Open files
			file, err := os.Open(reply.Job.Source[0])
			if err != nil {
				log.Fatalf("cannot open %v", reply.Job.Source)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", reply.Job.Source)
			}
			file.Close()

			// initialize array
			encoderList := make([]*json.Encoder, reply.NReduce)

			// Get pointer to all files and encoders
			for i := 0; i < reply.NReduce; i++ {
				tmpfile, err := ioutil.TempFile("", "mr-out-"+strconv.Itoa(i)+"*")
				if err != nil {
					log.Fatalf("cannot create temp file mr-out-%d", i)
				}
				doneJobArg.TempFiles = append(doneJobArg.TempFiles, tmpfile.Name())
				doneJobArg.TargetFiles = append(doneJobArg.TargetFiles, "mr-out-"+strconv.Itoa(i)+"-id-"+strconv.Itoa(reply.Job.ID))
				encoderList[i] = json.NewEncoder(tmpfile)
			}

			// Do the map and pass values into corresponding file
			for _, kv := range mapf(reply.Job.Source[0], string(content)) {
				index := ihash(kv.Key) % reply.NReduce
				err = encoderList[index].Encode(&kv)
				if err != nil {
					log.Fatalf("cannot encode %v in file %v", kv, index)
				}
			}
		} else {
			// Open files
			kva := []KeyValue{}
			for _, filename := range reply.Job.Source {
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				dec := json.NewDecoder(file)
				for {
					var kv KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					kva = append(kva, kv)
				}
				file.Close()
			}

			sort.Sort(ByKey(kva))

			ofile, err := ioutil.TempFile("", "mr-out-"+strconv.Itoa(reply.Job.ID-len(reply.Job.Source))+"-crash")
			if err != nil {
				log.Fatalf("cannot create temp file mr-out-%d", reply.Job.ID-len(reply.Job.Source))
			}
			doneJobArg.TempFiles = append(doneJobArg.TempFiles, ofile.Name())
			doneJobArg.TargetFiles = append(doneJobArg.TargetFiles, "mr-out-"+strconv.Itoa(reply.Job.ID-len(reply.Job.Source)))

			i := 0
			for i < len(kva) {
				j := i + 1
				for j < len(kva) && kva[j].Key == kva[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, kva[k].Value)
				}
				output := reducef(kva[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)
				i = j
			}
			ofile.Close()

			// Remove unnecessary files
			for _, filename := range reply.Job.Source {
				os.Remove(filename)
			}
		}
		call("Master.JobDoneHandler", &doneJobArg, &reply)
	}
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}
	return false
}
