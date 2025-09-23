package config

var RespNil = []byte("$-1\r\n")
var RespOk = []byte("+OK\r\n")

const (
	MaxConnection = 20000
	OpRead        = 0
	OpWrite       = 1
	CRLF          = "\r\n"
	WithScore     = "WITHSCORES"

	ERROR_WRONG_NUMBER_OF_ARGUMENTS         = "ERR wrong number of arguments for command"
	ERROR_SYNTAX                            = "ERR syntax error"
	ERROR_VALUE_NOT_INTEGER_OR_OUT_OF_RANGE = "ERR value is not an integer or out of range"
	ERROR_KEY_ALREADY_EXISTS                = "ERR key already exists"
	ERROR_WRONGTYPE                         = "WRONGTYPE Operation against a key holding the wrong kind of value"
	SCORE_MUST_BE_FLOAT                     = "Score must be floating point number"
	NO_DATA                                 = "No Data"
	KEY_NOT_EXIST                           = "key does not exist"
)
