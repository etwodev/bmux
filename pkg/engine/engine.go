package engine

import (
	"sync/atomic"
	"time"

	"github.com/etwodev/bmux/pkg/handler"
	"github.com/panjf2000/gnet/v2"
)

type ExtractLengthFunc[T any] func(c gnet.Conn, buf []byte) (headLen int, totalLen int)
type ExtractMsgIDFunc[T any] func(c gnet.Conn, head []byte) (msgID int)
type ContextFactoryFunc[T any] func() *T

type EngineWrapper[T any] struct {
	gnet.BuiltinEventEngine
	Engine            gnet.Engine
	ContextFactory    ContextFactoryFunc[T]
	ExtractLength     ExtractLengthFunc[T]
	ExtractMsgID      ExtractMsgIDFunc[T]
	LastIdleReset     time.Time
	ActiveConnections int64
	MaxConnections    int64
	HeadSize          int
	ReadTimeout       int
	WriteTimeout      int
	Handlers          map[int]handler.HandlerFunc
}

func (e *EngineWrapper[T]) OnBoot(eng gnet.Engine) gnet.Action {
	e.Engine = eng
	return gnet.None
}

func (e *EngineWrapper[T]) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	if atomic.LoadInt64(&e.ActiveConnections) >= e.MaxConnections {
		return nil, gnet.Close
	}
	atomic.AddInt64(&e.ActiveConnections, 1)
	c.SetContext(e.ContextFactory())
	return nil, gnet.None
}

func (e *EngineWrapper[T]) OnClose(c gnet.Conn, err error) gnet.Action {
	atomic.AddInt64(&e.ActiveConnections, -1)
	return gnet.None
}

func (e *EngineWrapper[T]) OnTraffic(c gnet.Conn) gnet.Action {
	var h handler.HandlerFunc
	var act gnet.Action
	var pkt [][]byte
	var buf []byte
	var err error
	var ok bool
	var ttl int
	var hd int

	if e.ReadTimeout > 0 {
		_ = c.SetReadDeadline(time.Now().Add(time.Duration(e.ReadTimeout) * time.Second))
	}

	buf, err = c.Next(e.HeadSize)
	if err != nil {
		goto close
	}

	hd, ttl = e.ExtractLength(c, buf)
	buf, err = c.Next(ttl)
	if err != nil {
		goto respond
	}

	h, ok = e.Handlers[e.ExtractMsgID(c, buf[:hd])]
	if !ok {
		goto respond
	}

	pkt, act = h(c, buf[hd:])
	c.Writev(pkt)
	return act
respond:
	return gnet.None
close:
	return gnet.Close
}
