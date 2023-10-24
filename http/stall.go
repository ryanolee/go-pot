package http

import (
	"bufio"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/ryan-pot/generator"
	"go.uber.org/zap"
)

type (
	HttpStaller struct {
		id           string
		generator    generator.Generator
		transferRate time.Duration

		running     bool
		runningLock sync.Mutex

		deregisterChan chan *HttpStaller
	}

	HttpStallerOptions struct {
		Id           string
		Generator    generator.Generator
		TransferRate time.Duration
		Request      *fiber.Ctx
	}
)

func NewHttpStaller(opts *HttpStallerOptions) *HttpStaller {
	if opts.TransferRate == 0 {
		opts.TransferRate = time.Millisecond * 2
	}

	return &HttpStaller{
		runningLock:  sync.Mutex{},
		running:      false,
		generator:    opts.Generator,
		transferRate: opts.TransferRate,
	}
}

// StallBuffer stalls the buffer by writing a chunk of data every N milliseconds
func (s *HttpStaller) StallContextBuffer(ctx *fiber.Ctx) error {
	connId := ctx.Context().ConnID()
	conn := ctx.Context().Conn()

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		logger := zap.L().Sugar()
		for {

			data := s.generator.GenerateChunk()
			logger.Infow("writing garbage data", "connId", connId, "transferRate", s.transferRate, "data", len(data))
			for i := 0; i < len(data); i++ {
				if !s.running {
					logger.Infow("connection closed", "connId", connId, "transferRate", s.transferRate)
					conn.Close()
					break
				}

				dataToWrite := []byte{}
				if string(data[i:i+1]) == "\\n" {
					dataToWrite = []byte("\\n")
					i++
				} else {
					dataToWrite = []byte{data[i]}
				}

				_, err := w.Write(dataToWrite)

				if err != nil {
					s.deregisterChan <- s
					s.Close()
					break
				}
				err = w.Flush()
				if err != nil {
					s.deregisterChan <- s
					s.Close()
					break
				}

				time.Sleep(s.transferRate)
			}
		}
	})

	return nil
}

func (s *HttpStaller) Close() {
	s.setRunning(false)
}

func (s *HttpStaller) setRunning(running bool) {
	s.runningLock.Lock()
	defer s.runningLock.Unlock()
	s.running = running
}
