package detect

import (
	"context"
	"net"
	"time"
)

type (
	// Connection that can be replayed
	rewindableConn struct {
		conn                 net.Conn
		buffer               []byte
		playBufferOnNextRead bool
	}
)

const (
	rewindBufferSize = 128
	pollInterval     = time.Millisecond * 100
)

func newRewindableConnFromConn(conn net.Conn) *rewindableConn {
	return &rewindableConn{
		conn:   conn,
		buffer: make([]byte, rewindBufferSize),
	}
}

// Custom method to rewind the connection
func (c *rewindableConn) Rewind() {
	c.playBufferOnNextRead = true
}

func (c *rewindableConn) Erase() {
	c.playBufferOnNextRead = false
	c.buffer = nil
}

func (c *rewindableConn) Close() error {
	return c.conn.Close()
}

func (c *rewindableConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *rewindableConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *rewindableConn) ReadWithTimeout(ctx context.Context, b []byte, timeout time.Duration) (int, error) {
	ticker := time.NewTicker(pollInterval)
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			n, err := c.Read(b)
			if (n > 0 && !allZeroes(b)) || err != nil {
				return n, err
			}
		case <-timeoutCtx.Done():
			return 0, nil
		}
	}
}

func (c *rewindableConn) Read(b []byte) (int, error) {
	// Replay buffer if needed
	if c.playBufferOnNextRead && c.buffer != nil {
		n := copy(b, c.buffer)

		// Once entire buffer is read, erase it
		if len(c.buffer) == 0 || allZeroes(c.buffer) {
			c.Erase()
			return n, nil
		}

		// Cut off part of the buffer that was read
		c.buffer = c.buffer[n:]

		return n, nil
	}

	// Read from the connection & handle error
	n, err := c.conn.Read(b)
	if err != nil {
		return n, err
	}

	// Copy back to buffer
	if c.buffer != nil {
		copy(c.buffer, b)
	}

	return n, nil
}

func (c *rewindableConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *rewindableConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *rewindableConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *rewindableConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func allZeroes(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
