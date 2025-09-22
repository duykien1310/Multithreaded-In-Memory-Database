package constant

const CRLF string = "\r\n"

var RespNil = []byte("$-1\r\n")
var RespOk = []byte("+OK\r\n")

const MaxConnection = 20000
const OpRead = 0
const OpWrite = 1
