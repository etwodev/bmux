package handler

import (
	"github.com/panjf2000/gnet/v2"
)

// HandlerFunc processes a message, returns zero or more packets to write and an action
type HandlerFunc func(conn gnet.Conn, body []byte) gnet.Action
