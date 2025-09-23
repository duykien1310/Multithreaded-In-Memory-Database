package command

import (
	"backend/internal/constant"
	"backend/internal/protocol/resp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (h *Handler) cmdZADD(args []string) []byte {
	if len(args) < 3 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZADD' command"), false)
	}
	key := args[0]
	scoreIndex := 1

	numScoreEleArgs := len(args) - scoreIndex
	if numScoreEleArgs%2 == 1 || numScoreEleArgs == 0 {
		return resp.Encode(errors.New(fmt.Sprintf("(error) Wrong number of (score, member) arg: %d", numScoreEleArgs)), false)
	}

	rs, err := h.datastore.ZADD(key, args[1:])
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Handler) cmdZSCORE(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZSCORE' command"), false)
	}

	key, member := args[0], args[1]
	score, exist, err := h.datastore.ZScore(key, member)
	if err != nil {
		return resp.Encode(err, false)
	}

	if !exist {
		return constant.RespNil
	}

	return resp.Encode(fmt.Sprintf("%f", score), false)
}

func (h *Handler) cmdZRANK(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZRANK' command"), false)
	}

	key, member := args[0], args[1]
	rank, exist, err := h.datastore.ZRank(key, member)
	if err != nil {
		return resp.Encode(err, false)
	}

	if !exist {
		return constant.RespNil
	}

	return resp.Encode(rank, false)
}

func (h *Handler) cmdZCARD(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZCARD' command"), false)
	}

	key := args[0]
	rs, err := h.datastore.ZCard(key)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Handler) cmdZRANGE(args []string) []byte {
	withScores := false
	if len(args) < 3 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZRANGE' command"), false)
	} else if len(args) == 4 {
		if strings.ToUpper(args[3]) != "WITHSCORES" {
			return resp.Encode(errors.New("ERR syntax error"), false)
		} else {
			withScores = true
		}
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	if withScores {
		rs, err := h.datastore.ZRangeWithScore(key, start, stop)
		if err != nil {
			return resp.Encode(err, false)
		}

		return resp.Encode(rs, false)
	}

	rs, err := h.datastore.ZRange(key, start, stop)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Handler) cmdZREM(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZREM' command"), false)
	}

	key := args[0]
	members := args[1:]

	rs, err := h.datastore.ZRem(key, members)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}
