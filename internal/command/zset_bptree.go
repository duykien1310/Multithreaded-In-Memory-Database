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

	count := 0
	for i := scoreIndex; i < len(args); i += 2 {
		member := args[i+1]
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return resp.Encode(errors.New("Score must be floating point number"), false)
		}
		ret := h.zset.ZADD(key, score, member)
		if ret != 1 {
			return resp.Encode(errors.New("error when adding element"), false)
		}
		count++
	}
	return resp.Encode(count, false)
}

func (h *Handler) cmdZSCORE(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZSCORE' command"), false)
	}

	key, member := args[0], args[1]
	score, exist := h.zset.ZScore(key, member)
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
	rank, exist := h.zset.ZRank(key, member)
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
	return resp.Encode(h.zset.ZCard(key), false)
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
		return resp.Encode(h.zset.ZRangeWithScore(key, start, stop), false)
	}

	return resp.Encode(h.zset.ZRange(key, start, stop), false)
}

func (h *Handler) cmdZREM(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ZREM' command"), false)
	}

	key := args[0]
	members := args[1:]
	deleted := 0

	for _, mem := range members {
		if h.zset.ZRem(key, mem) {
			deleted++
		}
	}

	return resp.Encode(deleted, false)
}
