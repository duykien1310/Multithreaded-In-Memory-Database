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
