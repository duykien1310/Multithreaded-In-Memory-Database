package resp

import (
	"backend/internal/entity"
	"errors"
	"strings"
)

func ParseCmd(data []byte) (*entity.Command, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	array := value.([]interface{})
	tokens := make([]string, len(array))
	for i := range tokens {
		tokens[i] = array[i].(string)
	}
	res := &entity.Command{Cmd: strings.ToUpper(tokens[0]), Args: tokens[1:]}
	return res, nil
}

// RESP format data => raw data
func Decode(data []byte) (interface{}, error) {
	res, _, err := DecodeOne(data)
	return res, err
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case ':':
		return decodeInt64(data)
	case '-':
		return decodeError(data)
	case '$':
		return decodeBulkString(data)
	case '*':
		return decodeArray(data)
	}
	return nil, 0, nil
}

// +OK\r\n => OK, 5
func decodeSimpleString(data []byte) (string, int, error) {
	pos := 1
	for data[pos] != '\r' {
		pos++
	}
	return string(data[1:pos]), pos + 2, nil
}

// :123\r\n => 123
func decodeInt64(data []byte) (int64, int, error) {
	var res int64 = 0
	pos := 1
	var sign int64 = 1
	if data[pos] == '-' {
		sign = -1
		pos++
	}
	if data[pos] == '+' {
		pos++
	}
	for data[pos] != '\r' {
		res = res*10 + int64(data[pos]-'0')
		pos++
	}

	return sign * res, pos + 2, nil
}

func decodeError(data []byte) (string, int, error) {
	return decodeSimpleString(data)
}

// $5\r\nhello\r\n => 5, 4
func readLen(data []byte) (int, int) {
	res, pos, _ := decodeInt64(data)
	return int(res), pos
}

// $5\r\nhello\r\n => "hello"
func decodeBulkString(data []byte) (string, int, error) {
	length, pos := readLen(data)
	return string(data[pos:(pos + length)]), pos + length + 2, nil
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n => {"hello", "world"}
func decodeArray(data []byte) (interface{}, int, error) {
	length, pos := readLen(data)
	var res []interface{} = make([]interface{}, length)

	for i := range res {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		res[i] = elem
		pos += delta
	}
	return res, pos, nil
}
