package wswrapper

import (
	"github.com/gorilla/websocket"
	"io"
	"net"
	"time"
)

// GORILLA TO NET.CONN CONVERTER STRUCT 🦍🛠️
type GorillaConn struct {
	ws     *websocket.Conn
	reader  io.Reader
}

func NewGorillaConn(ws *websocket.Conn) *GorillaConn {
	return &GorillaConn{ws: ws}
}

func (c *GorillaConn) Read(b []byte) (int, error) {
	for c.reader == nil {
		msgType, r, err := c.ws.NextReader()
		if err != nil {
			return 0, err
		}

		if msgType == websocket.BinaryMessage {
			c.reader = r
			break
		}
	}

	n, err := c.reader.Read(b)
	if err != io.EOF {
		return n, err
	}

	c.reader = nil // Clean reset for next network packet

	if n == 0 {
		return c.Read(b) // Fetch next frame if this one was empty whisper
	}

	return n, nil
}

func (c *GorillaConn) Write(b []byte) (int, error) {
	err := c.ws.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *GorillaConn) Close() error                       { return c.ws.Close() }
func (c *GorillaConn) LocalAddr() net.Addr                { return c.ws.LocalAddr() }
func (c *GorillaConn) RemoteAddr() net.Addr               { return c.ws.RemoteAddr() }
func (c *GorillaConn) SetDeadline(t time.Time) error      { return nil }
func (c *GorillaConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *GorillaConn) SetWriteDeadline(t time.Time) error { return nil }
