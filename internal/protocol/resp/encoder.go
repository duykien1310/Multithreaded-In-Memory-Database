package resp

import (
	"backend/internal/config"
	"bytes"
	"fmt"
)

func encodeString(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func encodeStringArray(sa []string) []byte {
	var buf bytes.Buffer
	buf.Grow(len(sa) * 32) // pre-allocate roughly

	fmt.Fprintf(&buf, "*%d\r\n", len(sa))
	for _, s := range sa {
		fmt.Fprintf(&buf, "$%d\r\n%s\r\n", len(s), s)
	}
	return buf.Bytes()
}

func encodeInt(v int) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", v))
}

func encodeUInt(v uint32) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", v))
}

func encodeIntArray(sa []int) []byte {
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, s := range sa {
		buf.Write(encodeInt(s))

	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(sa), buf.Bytes()))
}

func encodeUIntArray(sa []uint32) []byte {
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, s := range sa {
		buf.Write(encodeUInt(s))

	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(sa), buf.Bytes()))
}

// raw data => RESP format data
func Encode(value interface{}, isSimpleString bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimpleString {
			return []byte(fmt.Sprintf("+%s%s", v, config.CRLF))
		}
		return []byte(fmt.Sprintf("$%d%s%s%s", len(v), config.CRLF, v, config.CRLF))
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		return []byte(fmt.Sprintf(":%v\r\n", v))
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v))
	case []string:
		return encodeStringArray(v)
	case []int:
		return encodeIntArray(value.([]int))
	case []uint32:
		return encodeUIntArray(value.([]uint32))
	case [][]string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, sa := range value.([][]string) {
			buf.Write(encodeStringArray(sa))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(value.([][]string)), buf.Bytes()))
	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, x := range value.([]interface{}) {
			buf.Write(Encode(x, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(value.([]interface{})), buf.Bytes()))
	default:
		return config.RespNil
	}
}
