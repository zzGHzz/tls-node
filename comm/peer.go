package comm

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/zzGHzz/tls-node/logger"
)

var (
	FreqToRead = 1 * time.Second
	MaxMsgSize = 2048
)

// Peer handles the TLS connection to a peer
type Peer struct {
	ctx context.Context

	name string
	conn net.Conn

	msgCh chan []byte

	logger *slog.Logger
}

func NewPeer(ctx context.Context, name string, conn net.Conn, msgCh chan []byte) *Peer {
	peerlogger := logger.New(slog.LevelDebug).
		With("peer", name).
		With("local", conn.LocalAddr().String()).
		With("remote", conn.RemoteAddr().String())

	return &Peer{
		ctx:    ctx,
		name:   name,
		conn:   conn,
		msgCh:  msgCh,
		logger: peerlogger,
	}
}

// Listen listens for incoming messages from the peer. Once receving a message,
// it will send it to the message channel for handling.
func (p *Peer) Listen() {
	defer p.logger.Debug("stopped listening")

	p.logger.Debug("starting listening")

	ticker := time.NewTicker(FreqToRead)
	defer ticker.Stop()

	buf := make([]byte, MaxMsgSize)

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// p.conn.SetDeadline(time.Now().Add(ReadTimeout))
			n, err := p.conn.Read(buf)
			if err != nil { // when nothing read, n == 0 && err == EOF
				continue
			}

			p.msgCh <- buf[:n]
		}
	}
}

// Send sends a message to the peer
func (p *Peer) Send(msg Message) (int, error) {
	if b, err := msg.Serialize(); err == nil {
		return p.conn.Write(b)
	} else {
		return -1, err
	}
}

// Close closes the connection to the peer
func (p *Peer) Close() {
	defer p.logger.Debug("connection closed")

	if p.conn != nil {
		p.conn.Close()
	}
}

// Ping sends a ping message to the peer
func (p *Peer) Ping() error {
	msg := fmt.Sprintf("ping %s->%s", p.conn.LocalAddr().String(), p.conn.RemoteAddr().String())
	_, err := p.conn.Write([]byte(msg))
	return err
}