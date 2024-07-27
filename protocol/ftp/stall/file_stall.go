package stall

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/ryanolee/ryan-pot/config"
	stallLib "github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/generator/encoder"
)

// Acts as an IO.reader stream. It generates data using the generator and encoder and sends data in a requested chunk size
// It internally tracks how many bytes have been sent and gives an EOF once all req
type (
	FtpFileStaller struct {
		encoder   encoder.Encoder
		generator generator.Generator

		// Size of each chunk to send in calls to
		chunkSendSize int

		// Number of bytes generated. N.b this does not include the start, delimiter, padding or end, just the data
		bytesGenerated int

		// Current number of bytes sent
		bytesSent   int
		startedSend bool

		// Number of bytes staller should send
		bytesToSend int

		// Data for current chunk
		currentChunk     []byte
		currentChunkSize int
		currentChunkRead int

		// mutex sync.Mutex
		readMutex sync.Mutex

		// Staller Pool
		id             uint64
		groupId        string
		deregisterChan chan stallLib.Staller
		forcedEOF      bool
		closed         bool
	}

	NewFtpFileStallerArgs struct {
		// DI container for the ftp server
		Config *config.Config

		// Id of this specific staller
		Id uint64

		// Group identifier for the stalling client
		GroupId string

		// The associated encoder for the staller
		Encoder encoder.Encoder

		// The associated generator for the staller
		Generator generator.Generator

		// Number of bytes to send as part of the staller action
		BytesToSend int
	}
)

func NewFtpFileStall(args *NewFtpFileStallerArgs) *FtpFileStaller {
	return &FtpFileStaller{
		id:            args.Id,
		groupId:       args.GroupId,
		encoder:       args.Encoder,
		generator:     args.Generator,
		bytesToSend:   args.BytesToSend,
		chunkSendSize: args.Config.FtpServer.Transfer.ChunkSize,
		readMutex:     sync.Mutex{},
	}
}

// io.Reader interface implementation
func (f *FtpFileStaller) Read(p []byte) (n int, err error) {
	if f.forcedEOF {
		return 0, io.EOF
	}

	f.readMutex.Lock()
	defer f.readMutex.Unlock()

	// Check for sending start chunk and if so send it
	if f.bytesSent == 0 && !f.startedSend {
		f.startedSend = true
		// Send start chunk instead of regular generator chunk
		return f.sendData(f.start(), p)
	}

	// Check if we have an available buffered object
	if f.bytesGenerated == 0 {
		f.setNextChunk()
	} else if f.currentChunk == nil {
		// In cases where the current chunk in nil we have reached the end of
		// the last generated chunk we set the next chunk and send the delimiter
		f.setNextChunk()
		return f.sendData(f.delaminate(), p)
	}

	nextChunk := f.getBytesToSend()

	// Reset the current chunk if we have read all of it
	if f.currentChunkRead >= f.currentChunkSize {
		f.resetCurrentChunk()
	}

	return f.sendData(nextChunk, p)

}

// Sends first chunk of data
func (f *FtpFileStaller) start() []byte {
	return []byte(f.encoder.Start())
}

// Sends the delimiter
func (f *FtpFileStaller) delaminate() []byte {
	return []byte(f.encoder.Delimiter())
}

// Sends next chunk in stream
func (f *FtpFileStaller) sendData(source []byte, buffer []byte) (n int, err error) {
	// In the event we are done shutdown the data senders and close the streams
	if f.bytesSent >= f.bytesToSend {
		f.Halt()
		return 0, io.EOF
	}

	bytesCopied := copy(buffer, source)
	f.bytesSent += bytesCopied
	return bytesCopied, nil
}

func (f *FtpFileStaller) getBytesToSend() []byte {
	lowerBound := f.currentChunkRead
	upperBound := int(math.Min(float64(f.currentChunkRead+f.chunkSendSize), float64(f.currentChunkSize)))

	toSend := f.currentChunk[lowerBound:int(upperBound)]
	f.currentChunkRead += len(toSend)

	return toSend
}

func (f *FtpFileStaller) resetCurrentChunk() {
	f.currentChunkRead = 0
	f.currentChunk = nil
	f.currentChunkSize = 0
}

func (f *FtpFileStaller) setNextChunk() {

	// Generate the next chunk
	chunk := f.generator.Generate()

	// If generated chunk is larger than the remaining bytes to send fall back to sending last chunk
	if f.bytesSent+len(chunk) >= f.bytesToSend {
		f.setLastChunk()
		return
	}

	// Update current chunk with newly generated chunk
	f.currentChunkRead = 0
	f.currentChunk = chunk
	f.currentChunkSize = len(chunk)
	f.bytesGenerated += f.currentChunkSize
}

func (f *FtpFileStaller) setLastChunk() {
	end := f.generator.End()

	paddingToGenerate := f.bytesToSend - (f.bytesSent + len(end))

	// Generate spaces as padding
	padding := []byte(strings.Repeat(" ", int(math.Max(float64(paddingToGenerate), float64(0)))))

	finalChunk := append(padding, end...)

	f.currentChunk = finalChunk
	f.currentChunkSize = len(finalChunk)
	f.currentChunkRead = 0
}

func (f *FtpFileStaller) String() string {
	return fmt.Sprintf("FtpFileStaller{bytesToSend: %d, bytesSent: %d, bytesGenerated: %d, chunkSendSize: %d, currentChunkSize: %d, currentChunkRead: %d}",
		f.bytesToSend, f.bytesSent, f.bytesGenerated, f.chunkSendSize, f.currentChunkSize, f.currentChunkRead)
}

// Staller interface impl

func (f *FtpFileStaller) BindToPool(deregisterChan chan stallLib.Staller) {
	f.deregisterChan = deregisterChan
}

// Shuts down the staller instance and cleans up any resources
func (f *FtpFileStaller) Close() {
	f.forcedEOF = true
	f.closed = true
}

func (f *FtpFileStaller) Halt() {
	if !f.closed {
		f.deregisterChan <- f
	}
}

// Gets the group identifier for the staller
func (f *FtpFileStaller) GetGroupIdentifier() string {
	return f.groupId
}

// Gets the identifier for the staller
func (f *FtpFileStaller) GetIdentifier() uint64 {
	return f.id
}
