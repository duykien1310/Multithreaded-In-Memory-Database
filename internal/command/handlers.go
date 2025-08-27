package command

import (
	"backend/internal/datastore"
	"backend/internal/payload"
	"backend/internal/protocol/resp"
	"errors"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	if len(args) == 3 || len(args) > 4 {
		return resp.Encode(errors.New("ERR syntax error"), false)
	} else if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	key, val := args[0], []byte(args[1])

	// Options
	ttl := time.Duration(0)
	if len(args) >= 4 {
		opt := strings.ToUpper(args[2])
		if opt == "EX" {
			secs, _ := strconv.ParseInt(args[3], 10, 64)
			ttl = time.Duration(secs) * time.Second
		} else if opt == "PX" {
			ms, _ := strconv.ParseInt(args[3], 10, 64)
			ttl = time.Duration(ms) * time.Millisecond
		}
	}

	h.kv.Set(key, val, ttl)

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

func (h *Handler) cmdTTL(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	seconds := h.kv.TTL(args[0])

	return resp.Encode(strconv.Itoa(int(seconds)), true)
}

func (h *Handler) cmdPTTL(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'pttl' command"), false)
	}

	seconds := h.kv.PTTL(args[0])

	return resp.Encode(strconv.Itoa(int(seconds)), true)
}

func (h *Handler) cmdExpire(args []string) []byte {
	if len(args) > 2 {
		return resp.Encode(errors.New("ERR syntax error"), false)
	} else if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	sec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	if h.kv.Expire(args[0], sec) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}

func (h *Handler) cmdPExpire(args []string) []byte {
	if len(args) > 2 {
		return resp.Encode(errors.New("ERR syntax error"), false)
	} else if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	miliSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	if h.kv.PExpire(args[0], miliSec) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}

func (h *Handler) cmdPersist(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'persist' command"), false)
	}

	if h.kv.Persist(args[0]) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}
