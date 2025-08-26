package command

import (
	"backend/internal/datastore"
	"backend/internal/payload"
	"backend/internal/protocol/resp"
	"errors"
	"syscall"
)

type Handler struct {
	kv *datastore.KV
}

func NewHandler() *Handler {
	return &Handler{
		kv: datastore.NewKV(),
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
	default:
		res = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(connFd, res)
	return err
}

func (h *Handler) cmdPING(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		res = resp.Encode("PONG", true)
	} else {
		res = resp.Encode(args[0], false)
	}
	return res
}

func (h *Handler) cmdSET(args []string) []byte {
	if len(args) > 2 {
		return resp.Encode(errors.New("ERR syntax error"), false)
	} else if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	key, val := args[0], []byte(args[1])
	h.kv.Set(key, val)

	return resp.Encode("OK", true)
}

func (h *Handler) cmdGET(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	if val, ok := h.kv.Get(args[0]); ok {
		return resp.Encode(string(val), false)
	}

	return resp.Encode(nil, false)
}
