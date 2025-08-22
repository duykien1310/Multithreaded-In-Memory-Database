package command

import (
	"backend/internal/entity"
	"backend/internal/protocol/resp"
	"errors"
	"fmt"
	"syscall"
)

func HandleCmd(cmd *entity.Command, connFd int) error {
	var res []byte

	switch cmd.Cmd {
	case "PING":
		res = cmdPING(cmd.Args)
	default:
		res = []byte(fmt.Sprintf("-CMD NOT FOUND\r\n"))
	}
	_, err := syscall.Write(connFd, res)
	return err
}

func cmdPING(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		res = resp.Encode("PONG", true)
	} else {
		res = resp.Encode(args[0], false)
	}
	return res
}
