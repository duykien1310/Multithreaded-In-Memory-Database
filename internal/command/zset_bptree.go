package command

import (
	"backend/internal/constant"
	"backend/internal/protocol/resp"
	"errors"
	"fmt"
	"strconv"
)

func (h *Handler) cmdZADD(args []string) []byte {
	if len(args) < 3 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'ZADD' command"), false)
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
			return resp.Encode(errors.New("(error) Score must be floating point number"), false)
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
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'ZSCORE' command"), false)
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
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'ZRANK' command"), false)
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
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'ZCARD' command"), false)
	}

	key := args[0]
	return resp.Encode(h.zset.ZCard(key), false)
}
