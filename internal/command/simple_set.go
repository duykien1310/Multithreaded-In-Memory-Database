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

	return resp.Encode(h.simpleSet.SADD(key, members), false)
}

func (h *Handler) cmdSMembers(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'smembers' command"), false)
	}

	key := args[0]

	return resp.Encode(h.simpleSet.SMembers(key), false)
}

func (h *Handler) SIsMember(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'sismember' command"), false)
	}

	key, member := args[0], args[1]

	return resp.Encode(h.simpleSet.SIsMember(key, member), false)
}

func (h *Handler) SMIsMember(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'smismember' command"), false)
	}

	key, members := args[0], args[1:]

	return resp.Encode(h.simpleSet.SMIsMember(key, members), false)
}
