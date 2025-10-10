package worker

import (
	"backend/internal/datastore"
	"backend/internal/payload"
)

type Worker struct {
	id        int
	datastore *datastore.Datastore
	TaskCh    chan *payload.Task
}

func NewWorker(id int, bufferSize int, d *datastore.Datastore) *Worker {
	return &Worker{
		id:        id,
		datastore: d,
		TaskCh:    make(chan *payload.Task, bufferSize),
	}
}

func (w *Worker) Start() {
	for task := range w.TaskCh {
		w.HandleCmd(task)
	}
}

func (h *Worker) HandleCmd(task *payload.Task) {
	var res []byte

	switch task.Command.Cmd {
	case "PING":
		res = h.cmdPING(task.Command.Args)
	case "SET":
		res = h.cmdSET(task.Command.Args)
	case "GET":
		res = h.cmdGET(task.Command.Args)
	case "TTL":
		res = h.cmdTTL(task.Command.Args)
	case "PTTL":
		res = h.cmdPTTL(task.Command.Args)
	case "EXPIRE":
		res = h.cmdExpire(task.Command.Args)
	case "PEXPIRE":
		res = h.cmdPExpire(task.Command.Args)
	case "PERSIST":
		res = h.cmdPersist(task.Command.Args)
	case "EXISTS":
		res = h.cmdExists(task.Command.Args)
	case "DEL":
		res = h.cmdDel(task.Command.Args)

	// Simple Set
	case "SADD":
		res = h.cmdSADD(task.Command.Args)
	case "SMEMBERS":
		res = h.cmdSMembers(task.Command.Args)
	case "SISMEMBER":
		res = h.cmdSIsMember(task.Command.Args)
	case "SMISMEMBER":
		res = h.cmdSMIsMember(task.Command.Args)

	// Sorted Set
	case "ZADD":
		res = h.cmdZADD(task.Command.Args)
	case "ZSCORE":
		res = h.cmdZSCORE(task.Command.Args)
	case "ZRANK":
		res = h.cmdZRANK(task.Command.Args)
	case "ZCARD":
		res = h.cmdZCARD(task.Command.Args)
	case "ZRANGE":
		res = h.cmdZRANGE(task.Command.Args)
	case "ZREM":
		res = h.cmdZREM(task.Command.Args)

	// CMS
	case "CMS.INITBYDIM":
		res = h.cmdCMSINITBYDIM(task.Command.Args)
	case "CMS.INITBYPROB":
		res = h.cmdCMSINITBYPROB(task.Command.Args)
	case "CMS.INCRBY":
		res = h.cmdCMSINCRBY(task.Command.Args)
	case "CMS.QUERY":
		res = h.cmdCMSQUERY(task.Command.Args)
	case "CMS.INFO":
		res = h.cmdINFO(task.Command.Args)

	default:
		res = []byte("-CMD NOT FOUND\r\n")
	}
	// _, err := syscall.Write(task.ConnFd, res)
	// return err

	task.ReplyCh <- res
}
