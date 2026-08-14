package internal

import (
	"errors"
	"io"

	"github.com/guestin/mob/mio"
)

type ReplayBuffer struct {
	raw          io.ReadCloser
	rawReadError error
	eofReach     bool
	replayFlag   bool
	data         []byte
	rIdx         int
}

func isEOF(err error) bool {
	return errors.Is(io.EOF, err)
}

func NewReplayBuffer(raw io.ReadCloser) io.ReadSeekCloser {
	return &ReplayBuffer{
		raw:          raw,
		rawReadError: nil,
	}
}

func (this *ReplayBuffer) len() int {
	return len(this.data) - this.rIdx
}

func (this *ReplayBuffer) record(p []byte) (int, error) {
	this.data = append(this.data, p...)
	this.rIdx += len(p)
	return len(p), nil
}

func (this *ReplayBuffer) replayRead(p []byte) (int, error) {
	bufLen := this.len()
	if bufLen == 0 {
		return 0, io.EOF
	}
	expectN := mio.MinInt(bufLen, len(p))
	copy(p, this.data[this.rIdx:this.rIdx+expectN])
	this.rIdx += expectN
	return expectN, nil
}

func (this *ReplayBuffer) Read(p []byte) (int, error) {
	if this.eofReach || this.replayFlag {
		return this.replayRead(p)
	}
	if this.rawReadError == nil {
		n, err := this.raw.Read(p)
		eofFlag := isEOF(err)
		if err == nil || eofFlag {
			// record for replay
			_, _ = this.record(p[:n])
			if eofFlag {
				this.eofReach = true
			}
		} else {
			this.rawReadError = err
		}
		return n, err
	}
	return 0, this.rawReadError
}

func (this *ReplayBuffer) Close() error {
	this.data = make([]byte, 0)
	return this.raw.Close()
}

func (this *ReplayBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		this.rIdx += int(offset)
	case io.SeekStart:
		this.rIdx = int(offset)
		this.replayFlag = true
	case io.SeekEnd:
		this.rIdx = len(this.data) - int(offset)
	}
	return int64(this.rIdx), nil
}
