package stall

import (
	"bufio"
	"context"
	"errors"
	"math"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"go.uber.org/zap"
)

const (
	// Rate at which staller will report on wasted time to the given telemetry instance
	StallerReportInterval = time.Second * 30
)

type (
	HttpStaller struct {
		id           uint64
		ipAddress    string
		generator    generator.Generator
		transferRate time.Duration
		ticker       *time.Ticker
		timeout      time.Duration
		startTime    time.Time
		endTime      time.Time
		onTimeout    func(*HttpStaller)
		onClose      func(*HttpStaller)

		running     bool
		runningLock sync.Mutex

		deregisterChan chan *HttpStaller

		telemetryTicker *time.Ticker
		telemetry       *metrics.Telemetry
	}

	HttpStallerOptions struct {
		Generator    generator.Generator
		TransferRate time.Duration
		Request      *fiber.Ctx
		Timeout      time.Duration
		OnTimeout    func(*HttpStaller)
		OnClose      func(*HttpStaller)
		Telemetry    *metrics.Telemetry
	}
)

func NewHttpStaller(opts *HttpStallerOptions) *HttpStaller {
	if opts.TransferRate == 0 {
		opts.TransferRate = time.Millisecond * 75
	}

	if opts.Timeout == 0 {
		opts.Timeout = time.Second * 10
	}

	if opts.OnClose == nil {
		opts.OnClose = func(_ *HttpStaller) {}
	}

	if opts.OnTimeout == nil {
		opts.OnTimeout = func(_ *HttpStaller) {}
	}

	return &HttpStaller{
		runningLock:  sync.Mutex{},
		running:      true,
		generator:    opts.Generator,
		transferRate: opts.TransferRate,
		timeout:      opts.Timeout,
		ipAddress:    opts.Request.IP(),
		id:           opts.Request.Context().ConnID(),
		onTimeout:    opts.OnTimeout,
		onClose:      opts.OnClose,
		telemetry:    opts.Telemetry,
	}
}

func (s *HttpStaller) BindPool(deregisterChan chan *HttpStaller) {
	s.deregisterChan = deregisterChan
}

// StallBuffer stalls the buffer by writing a chunk of data every N milliseconds
func (s *HttpStaller) StallContextBuffer(ctx *fiber.Ctx) error {
	conn := ctx.Context().Conn()
	s.startTime = time.Now()

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		s.ticker = time.NewTicker(s.transferRate)
		s.telemetryTicker = time.NewTicker(StallerReportInterval)
		logger := zap.L().Sugar()

		closeContext := context.Background()
		closeContext, cancelContext := context.WithTimeout(closeContext, s.timeout)

		continueWriting, err := s.PushDataToClient(closeContext, w, s.generator.Start())
		if !continueWriting {
			logger.Error("staller closed during header write", "connId", s.id, "err", err)
			s.Halt(conn)
			cancelContext()
			return
		}

		for {
			if errors.Is(closeContext.Err(), context.DeadlineExceeded) {
				// Flush the rest of the data to the client in the case we are closing
				w.Write(s.generator.End())
				w.Flush()
				s.Halt(conn)
				s.handleClose()
				cancelContext()
				return
			}

			data := s.generator.GenerateChunk()

			continueWriting, _ := s.PushDataToClient(closeContext, w, data)

			if !continueWriting {
				time.Sleep(time.Millisecond * 500)
				s.Halt(conn)
				cancelContext()
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
				s.handleTimeout()
				return false, err
			}

			if err := w.Flush(); err != nil {
				s.handleTimeout()
				return false, err
			}
		case <-s.telemetryTicker.C:
			if s.telemetry == nil {
				continue
			}

			s.telemetry.TrackWastedTime(StallerReportInterval)
		case <-ctx.Done():
			// Flush the rest of the data to the client in the case we are closing
			w.Write(data[i:])
			w.Flush()
			w.Write(s.generator.End())
			w.Flush()
			s.handleClose()
			return false, nil
		}
	}

	return true, nil
}

func (s *HttpStaller) Halt(conn net.Conn) {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.telemetryTicker != nil {
		s.telemetryTicker.Stop()
	}
	s.deregisterChan <- s
	s.Close()
	conn.Close()
}

func (s *HttpStaller) handleTimeout() {
	s.endTime = time.Now()
	go s.onTimeout(s)
	if s.telemetry != nil {
		s.telemetry.TrackWastedTime(s.GetRemainingTimeToReport())
	}
}

func (s *HttpStaller) handleClose() {
	s.endTime = time.Now()
	go s.onClose(s)
	if s.telemetry != nil {
		s.telemetry.TrackWastedTime(s.GetRemainingTimeToReport())
	}
}

func (s *HttpStaller) GetElapsedTime() time.Duration {
	return s.endTime.Sub(s.startTime)
}

func (s *HttpStaller) GetRemainingTimeToReport() time.Duration {
	return time.Duration(math.Mod(float64(s.GetElapsedTime()), float64(StallerReportInterval)))
}

func (s *HttpStaller) Close() {
	s.setRunning(false)
}

func (s *HttpStaller) setRunning(running bool) {
	s.runningLock.Lock()
	defer s.runningLock.Unlock()
	s.running = running
}
