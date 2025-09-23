package config

import "errors"

var ErrWrongNumberArguments = errors.New(ERROR_WRONG_NUMBER_OF_ARGUMENTS)
var ErrSyntaxError = errors.New(ERROR_SYNTAX)
var ErrValueNotIntegerOrOutOfRange = errors.New(ERROR_VALUE_NOT_INTEGER_OR_OUT_OF_RANGE)
var ErrKeyAlreadyExists = errors.New(ERROR_KEY_ALREADY_EXISTS)
var ErrWrongType = errors.New(ERROR_WRONGTYPE)
var ErrScoreIsNotFloat = errors.New(SCORE_MUST_BE_FLOAT)
var ErrNoData = errors.New(NO_DATA)
var ErrKeyNotExist = errors.New(KEY_NOT_EXIST)
