package worker

import (
	"backend/internal/config"
	"backend/internal/protocol/resp"
	"fmt"
	"strconv"
	"strings"
)

func (h *Worker) cmdZADD(args []string) []byte {
	if len(args) < 3 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}
	key := args[0]
	scoreIndex := 1

	numScoreEleArgs := len(args) - scoreIndex
	if numScoreEleArgs%2 == 1 || numScoreEleArgs == 0 {
		return resp.Encode(fmt.Errorf("(error) Wrong number of (score, member) arg: %d", numScoreEleArgs), false)
	}

	rs, err := h.datastore.ZADD(key, args[1:])
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Worker) cmdZSCORE(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	key, member := args[0], args[1]
	score, exist, err := h.datastore.ZScore(key, member)
	if err != nil {
		return resp.Encode(err, false)
	}

	if !exist {
		return config.RespNil
	}

	return resp.Encode(fmt.Sprintf("%f", score), false)
}

func (h *Worker) cmdZRANK(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	key, member := args[0], args[1]
	rank, exist, err := h.datastore.ZRank(key, member)
	if err != nil {
		return resp.Encode(err, false)
	}

	if !exist {
		return config.RespNil
	}

	return resp.Encode(rank, false)
}

func (h *Worker) cmdZCARD(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	key := args[0]
	rs, err := h.datastore.ZCard(key)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}

func (h *Worker) cmdZRANGE(args []string) []byte {
	if len(args) < 3 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	withScores := false
	if len(args) >= 4 {
		if strings.EqualFold(args[3], "WITHSCORES") {
			withScores = true
		} else {
			return resp.Encode(config.ErrSyntaxError, false)
		}
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}

	var res []string
	if withScores {
		res, err = h.datastore.ZRangeWithScore(key, start, stop)
	} else {
		res, err = h.datastore.ZRange(key, start, stop)
	}
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(res, false)
}

func (h *Worker) cmdZREVRANGE(args []string) []byte {
	if len(args) < 3 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	withScores := false
	if len(args) >= 4 {
		if strings.EqualFold(args[3], "WITHSCORES") {
			withScores = true
		} else {
			return resp.Encode(config.ErrSyntaxError, false)
		}
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.Encode(config.ErrValueNotIntegerOrOutOfRange, false)
	}

	var res []string
	if withScores {
		res, err = h.datastore.ZRevRangeWithScore(key, start, stop)
	} else {
		res, err = h.datastore.ZRevRange(key, start, stop)
	}
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(res, false)
}

func (h *Worker) cmdZREM(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(config.ErrWrongNumberArguments, false)
	}

	key := args[0]
	members := args[1:]

	rs, err := h.datastore.ZRem(key, members)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(rs, false)
}
