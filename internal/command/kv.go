package command

import (
	"backend/internal/config"
	"backend/internal/protocol/resp"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) cmdPING(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
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
		return resp.Encode(config.ErrSyntaxError, false)
	} else if len(args) < 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	key, val := args[0], args[1]

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

	h.datastore.Set(key, val, ttl)

	return resp.Encode("OK", true)
}

func (h *Handler) cmdGET(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	val, ok, err := h.datastore.Get(args[0])
	if err != nil {
		return resp.Encode(err, false)
	}

	if !ok {
		return config.RespNil
	}

	return resp.Encode(val, false)
}

func (h *Handler) cmdTTL(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	seconds := h.datastore.TTL(args[0])

	return resp.Encode(strconv.Itoa(int(seconds)), true)
}

func (h *Handler) cmdPTTL(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	seconds := h.datastore.PTTL(args[0])

	return resp.Encode(strconv.Itoa(int(seconds)), true)
}

func (h *Handler) cmdExpire(args []string) []byte {
	if len(args) > 2 {
		return resp.Encode(config.ErrSyntaxError, false)
	} else if len(args) < 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	sec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}

	if h.datastore.Expire(args[0], sec) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}

func (h *Handler) cmdPExpire(args []string) []byte {
	if len(args) > 2 {
		return resp.Encode(config.ErrSyntaxError, false)
	} else if len(args) < 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	miliSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}

	if h.datastore.PExpire(args[0], miliSec) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}

func (h *Handler) cmdPersist(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	if h.datastore.Persist(args[0]) {
		return resp.Encode(strconv.Itoa(1), true)
	}

	return resp.Encode(strconv.Itoa(0), true)
}

func (h *Handler) cmdExists(args []string) []byte {
	if len(args) < 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	count := h.datastore.Exists(args)

	return resp.Encode(strconv.Itoa(count), true)
}

func (h *Handler) cmdDel(args []string) []byte {
	if len(args) < 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	count := h.datastore.Del(args)

	return resp.Encode(strconv.Itoa(count), true)
}
