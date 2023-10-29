package stall

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/ryan-pot/generator"
	"go.uber.org/zap"
)

type (
	HttpStaller struct {
		id           uint64
		ipAddress    string
		generator    generator.Generator
		transferRate time.Duration
		ticker 		 *time.Ticker

		running     bool
		runningLock sync.Mutex

		deregisterChan chan *HttpStaller
	}

	HttpStallerOptions struct {
		Generator    generator.Generator
		TransferRate time.Duration
		Request      *fiber.Ctx
	}
)

func NewHttpStaller(opts *HttpStallerOptions) *HttpStaller {
	if opts.TransferRate == 0 {
		opts.TransferRate = time.Millisecond * 75
	}

	return &HttpStaller{
		runningLock:  sync.Mutex{},
		running:      true,
		generator:    opts.Generator,
		transferRate: opts.TransferRate,
		ipAddress:    opts.Request.IP(),
		id:           opts.Request.Context().ConnID(),
	}
}

func (s *HttpStaller) BindPool(deregisterChan chan *HttpStaller) {
	s.deregisterChan = deregisterChan
}

// StallBuffer stalls the buffer by writing a chunk of data every N milliseconds
func (s *HttpStaller) StallContextBuffer(ctx *fiber.Ctx) error {
	conn := ctx.Context().Conn()

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		s.ticker = time.NewTicker(s.transferRate)
		logger := zap.L().Sugar()

		closeContext := context.Background()
		closeContext, _ = context.WithTimeout(closeContext, time.Second* 10)

		continueWriting, err := s.PushDataToClient(closeContext, w, s.generator.Start())
		if !continueWriting {
			logger.Error("staller closed during header write", "connId", s.id, "err", err)
			s.Halt(conn)
			return
		}
		
		for {
			if errors.Is(closeContext.Err(), context.DeadlineExceeded) {
				logger.Info("staller closed", "connId", s.id)
				// Flush the rest of the data to the client in the case we are closing
				w.Write(s.generator.End())
				w.Flush()

				s.Halt(conn)
				return
			}

			data := s.generator.GenerateChunk()
			logger.Infow("writing garbage data", "connId", s.id, "transferRate", s.transferRate, "data", len(data))
			
			continueWriting, err := s.PushDataToClient(closeContext, w, data)
			
			
			if !continueWriting {
				time.Sleep(time.Millisecond * 1000)
				logger.Infow("staller closed", "connId", s.id, "error", err)
				s.Halt(conn)
				return
			}
			s.PushDataToClient(closeContext, w, s.generator.ChunkSeparator())
		}
	})
	return nil
}

func (s *HttpStaller) PushDataToClient(ctx context.Context, w *bufio.Writer, data []byte) (bool, error) {

	for i := 0; i < len(data); i++ {
		select {
			case <-s.ticker.C:
				if !s.running {
					return false, nil
				}

				dataToWrite := []byte{}
				if string(data[i:i+1]) == "\\n" {
					dataToWrite = []byte("\\n")
					i++
				} else {
					dataToWrite = []byte{data[i]}
				}
				
				if _, err := w.Write(dataToWrite); err != nil {
					return false, err
				}

				if err := w.Flush(); err != nil {
					return false, err
				}
			case <-ctx.Done():
				// Flush the rest of the data to the client in the case we are closing
				zap.L().Sugar().Infow("staller closed flushing remaining data", "connId", s.id, "data", len(data[i:]))
				w.Write(data[i:])
				w.Flush()
				w.Write(s.generator.End())
				w.Flush()
				return false, nil
		}
	}

	return true, nil
}

func (s *HttpStaller) Halt(conn net.Conn) {
	s.deregisterChan <- s
	s.Close()
	conn.Close()
}

func (s *HttpStaller) Close() {
	s.setRunning(false)
}

func (s *HttpStaller) setRunning(running bool) {
	s.runningLock.Lock()
	defer s.runningLock.Unlock()
	s.running = running
}
