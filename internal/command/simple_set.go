package command

import (
	"backend/internal/protocol/resp"
	"errors"
)

func (h *Handler) cmdSADD(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'sadd' command"), false)
	}

	key, members := args[0], args[1:]
	countAdded, err := h.datastore.SADD(key, members)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(countAdded, false)
}

func (h *Handler) cmdSMembers(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'smembers' command"), false)
	}

	key := args[0]
	rs, err := h.datastore.SMembers(key)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Handler) cmdSIsMember(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'sismember' command"), false)
	}

	key, member := args[0], args[1]
	rs, err := h.datastore.SIsMember(key, member)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Handler) cmdSMIsMember(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'smismember' command"), false)
	}

	key, members := args[0], args[1:]
	rs, err := h.datastore.SMIsMember(key, members)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}
