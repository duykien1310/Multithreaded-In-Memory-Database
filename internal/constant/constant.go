package constant

import "time"

var RespNil = []byte("$-1\r\n")
var RespOk = []byte("+OK\r\n")
var TtlKeyNotExist = []byte(":-2\r\n")
var TtlKeyExistNoExpire = []byte(":-1\r\n")

const ActiveExpireFrequency = 100 * time.Millisecond
const ActiveExpireSampleSize = 20
const ActiveExpireThreshold = 0.1
const MaxConnection = 1000

const OpRead = 0
const OpWrite = 1
