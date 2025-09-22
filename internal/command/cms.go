package command

import (
	"backend/internal/constant"
	"backend/internal/protocol/resp"
	"errors"
	"fmt"
	"math"
	"strconv"
)

func (h *Handler) cmdCMSINITBYDIM(args []string) []byte {
	if len(args) != 3 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INITBYDIM' command"), false)
	}
	key := args[0]
	width, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return resp.Encode(fmt.Errorf("width must be a integer number %s", args[1]), false)
	}
	height, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		return resp.Encode(fmt.Errorf("height must be a integer number %s", args[1]), false)
	}

	if !h.cms.CreateCMS(key, uint32(width), uint32(height)) {
		return resp.Encode(errors.New("CMS: key already exists"), false)
	}

	return constant.RespOk
}

func (h *Handler) cmdCMSINITBYPROB(args []string) []byte {
	if len(args) != 3 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INITBYPROB' command"), false)
	}
	key := args[0]
	errRate, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return resp.Encode(fmt.Errorf("errRate must be a floating point number %s", args[1]), false)
	}
	if errRate >= 1 || errRate <= 0 {
		return resp.Encode(errors.New("CMS: invalid overestimation value"), false)
	}
	probability, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.Encode(fmt.Errorf("probability must be a floating poit number %s", args[2]), false)
	}
	if probability >= 1 || probability <= 0 {
		return resp.Encode(errors.New("CMS: invalid prob value"), false)
	}

	if !h.cms.CreateCMSByProb(key, errRate, probability) {
		return resp.Encode(errors.New("CMS: key already exists"), false)
	}

	return constant.RespOk
}

func (h *Handler) cmdCMSINCRBY(args []string) []byte {
	if len(args) < 3 || len(args)%2 == 0 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INCBY' command"), false)
	}

	key := args[0]
	var res []string
	for i := 1; i < len(args); i += 2 {
		item := args[i]
		value, err := strconv.ParseUint(args[i+1], 10, 32)
		if err != nil {
			return resp.Encode(fmt.Errorf("increment must be a non negative integer number %s", args[1]), false)
		}
		count, err := h.cms.IncrBy(key, item, uint32(value))
		if err != nil {
			return resp.Encode(err, false)
		}
		if count == math.MaxUint32 {
			res = append(res, "CMS: INCRBY overflow")
			continue
		}
		res = append(res, fmt.Sprintf("%d", count))
	}
	return resp.Encode(res, false)
}

func (h *Handler) cmdCMSQUERY(args []string) []byte {
	if len(args) < 2 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.QUERY' command"), false)
	}

	key := args[0]
	items := args[1:]
	res, err := h.cms.Query(key, items)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode(res, false)
}

func (h *Handler) cmdINFO(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INFO' command"), false)
	}

	key := args[0]
	w, d, err := h.cms.Info(key)
	if err != nil {
		return resp.Encode(err, false)
	}

	return resp.Encode([]any{"width", w, "depth", d}, false)
}
