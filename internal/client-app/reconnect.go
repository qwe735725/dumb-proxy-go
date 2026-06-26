package clientapp

import (
	"sync"
	"sync/atomic"
	"time"

	"dumb-proxy-go/pkg/wswrapper"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
)

var MinReconnectDelay = 1 * time.Second

type MasterConn struct {
	wsUrl string

	reconnecting    atomic.Bool
	rw              sync.RWMutex
	wsConn          *websocket.Conn
	yamuxConn       *yamux.Session
	lastReconnectAt time.Time
}

func NewMasterConn(wsUrl string) *MasterConn {
	return &MasterConn{
		wsUrl: wsUrl,
	}
}

func (m *MasterConn) TriggerReconnect() bool {
	if !m.reconnecting.CompareAndSwap(false, true) {
		return false
	}

	go func() error {
		defer m.reconnecting.Store(false)

		m.rw.Lock()
		defer m.rw.Unlock()

		if time.Since(m.lastReconnectAt) < MinReconnectDelay {
			return errors.New("too many reconnects")
		}

		wsConn, _, err := websocket.DefaultDialer.Dial(m.wsUrl, nil)
		if err != nil {
			return err
		}

		yamuxConn, err := yamux.Client(wswrapper.NewGorillaConn(wsConn), nil)
		if err != nil {
			wsConn.Close()
			return err
		}

		m.wsConn = wsConn
		m.yamuxConn = yamuxConn
		m.lastReconnectAt = time.Now()

		return nil
	}()

	return true
}

func (m *MasterConn) WsConn() *websocket.Conn {
	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.wsConn
}

func (m *MasterConn) YamuxConn() *yamux.Session {
	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.yamuxConn
}
