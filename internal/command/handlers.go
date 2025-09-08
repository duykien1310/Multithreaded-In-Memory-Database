package command

import (
	"backend/internal/datastore"
	"backend/internal/payload"
	"syscall"
)

type Handler struct {
	kv        *datastore.KV
	simpleSet *datastore.SimpleSet
}

func NewHandler() *Handler {
	return &Handler{
		kv:        datastore.NewKV(),
		simpleSet: datastore.NewSimpleSet(),
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
	case "SADD":
		res = h.cmdSADD(cmd.Args)
	case "SMEMBERS":
		res = h.cmdSMembers(cmd.Args)
	case "SISMEMBER":
		res = h.SIsMember(cmd.Args)
	case "SMISMEMBER":
		res = h.SMIsMember(cmd.Args)
	default:
		res = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(connFd, res)
	return err
}
