package engine

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/etwodev/bmux/pkg/handler"
	"github.com/panjf2000/gnet/v2"
	"github.com/rs/zerolog"
)

var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: "2006-01-02T15:04:05",
}).With().Timestamp().Str("Group", "bmux-engine").Logger()

type ExtractLengthFunc[T any] func(c gnet.Conn, buf []byte) (headLen int, totalLen int)
type ExtractMsgIDFunc[T any] func(c gnet.Conn, head []byte, body []byte) (msgID int)
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
	var buf []byte
	var err error
	var ok bool
	var ttl int
	var hd int

	buf, err = c.Next(e.HeadSize)
	if err != nil {
		log.Warn().
			Err(err).
			Str("remote", c.RemoteAddr().String()).
			Msg("failed to read header from connection")

		goto respond
	}

	hd, ttl = e.ExtractLength(c, buf)
	buf, err = c.Next(ttl)
	if err != nil {
		log.Warn().
			Err(err).
			Str("remote", c.RemoteAddr().String()).
			Int("expected", ttl).
			Msg("failed to read full payload from connection")

		goto respond
	}

	h, ok = e.Handlers[e.ExtractMsgID(c, buf[:hd], buf[hd:])]
	if !ok {
		log.Warn().
			Str("remote", c.RemoteAddr().String()).
			Msg("no handler registered for message")

		goto respond
	}

	return h(c, buf[hd:])
respond:
	return gnet.None
}
