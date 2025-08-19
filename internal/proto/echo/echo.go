package echo

import (
	"bytes"
)

// Simple line-based protocol handler demonstrating pluggability.
// Echoes each full line back and keeps the connection open.

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) OnRead(fd int, in []byte) (out []byte, closeAfterWrite bool, err error) {
	// Find last full line (\n)
	idx := bytes.LastIndexByte(in, '\n')
	if idx < 0 {
		return nil, false, nil
	} // incomplete line
	// Echo back the complete portion.
	return append([]byte(nil), in[:idx+1]...), false, nil
}
