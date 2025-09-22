package command

import (
	"backend/internal/datastore"
	"backend/internal/payload"
	"syscall"
)

type Handler struct {
	kv        *datastore.KV
	simpleSet *datastore.SimpleSet
	zset      *datastore.ZSetBPTree
	cms       *datastore.CMS
}

func NewHandler() *Handler {
	return &Handler{
		kv:        datastore.NewKV(),
		simpleSet: datastore.NewSimpleSet(),
		zset:      datastore.NewZSetBPTree(),
		cms:       datastore.NewCMS(),
	}
}

func (h *Handler) HandleCmd(cmd *payload.Command, connFd int) error {
	var res []byte

	switch cmd.Cmd {
	case "PING":
		res = h.cmdPING(cmd.Args)
	case "SET":
		res = h.cmdSET(cmd.Args)
	case "GET":
		res = h.cmdGET(cmd.Args)
	case "TTL":
		res = h.cmdTTL(cmd.Args)
	case "PTTL":
		res = h.cmdPTTL(cmd.Args)
	case "EXPIRE":
		res = h.cmdExpire(cmd.Args)
	case "PEXPIRE":
		res = h.cmdPExpire(cmd.Args)
	case "PERSIST":
		res = h.cmdPersist(cmd.Args)
	case "EXISTS":
		res = h.cmdExists(cmd.Args)
	case "DEL":
		res = h.cmdDel(cmd.Args)

	// Simple Set
	case "SADD":
		res = h.cmdSADD(cmd.Args)
	case "SMEMBERS":
		res = h.cmdSMembers(cmd.Args)
	case "SISMEMBER":
		res = h.cmdSIsMember(cmd.Args)
	case "SMISMEMBER":
		res = h.cmdSMIsMember(cmd.Args)

	// Sorted Set
	case "ZADD":
		res = h.cmdZADD(cmd.Args)
	case "ZSCORE":
		res = h.cmdZSCORE(cmd.Args)
	case "ZRANK":
		res = h.cmdZRANK(cmd.Args)
	case "ZCARD":
		res = h.cmdZCARD(cmd.Args)
	case "ZRANGE":
		res = h.cmdZRANGE(cmd.Args)
	case "ZREM":
		res = h.cmdZREM(cmd.Args)

	// CMS
	case "CMS.INITBYDIM":
		res = h.cmdCMSINITBYDIM(cmd.Args)
	case "CMS.INITBYPROB":
		res = h.cmdCMSINITBYPROB(cmd.Args)
	case "CMS.INCRBY":
		res = h.cmdCMSINCRBY(cmd.Args)
	case "CMS.QUERY":
		res = h.cmdCMSQUERY(cmd.Args)
	case "CMS.INFO":
		res = h.cmdINFO(cmd.Args)

	default:
		res = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(connFd, res)
	return err
}
