package http

import (
	"bufio"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/ryan-pot/generator"
)
type (
	HttpStaller struct {
		generator generator.Generator
		transferRate time.Duration
	}

	HttpStallerOptions struct {
		Generator generator.Generator
		TransferRate time.Duration
		Request *fiber.Ctx
	}
)

func NewHttpStaller(opts *HttpStallerOptions) *HttpStaller {
	if opts.TransferRate == 0 {
		opts.TransferRate = time.Millisecond * 2
	}

	return &HttpStaller{
		generator: opts.Generator,
		transferRate: opts.TransferRate,
	}
}

// StallBuffer stalls the buffer by writing a chunk of data every 250ms
func (s *HttpStaller) StallContextBuffer(ctx *fiber.Ctx) error {
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			data := s.generator.GenerateChunk()
			for i:=0; i<len(data); i++ {
				dataToWrite := []byte{}
				if(string(data[i:i+1]) == "\\n") {
					dataToWrite = []byte("\\n")
					i++
				} else {
					dataToWrite = []byte{data[i]}
				}
				_, err := w.Write(dataToWrite)

				if err != nil {
					break
				}
				err = w.Flush()
				if err != nil {
					break
				}
				time.Sleep(s.transferRate)
			}
		}
	})

	return nil
}
	