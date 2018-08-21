package websocket

import (
	"bufio"
	"errors"
	"sync"
)

// Pool is a general-purpose interface for pooling.
// Pool is implemented by sync.Pool.
type Pool interface {
	// Get returns a new value.
	// Must never return nil.
	Get() interface{}

	// Put saves a value if possible.
	Put(interface{})
}

// NewBufferPool creates a Pool of byte slices with len of size.
func NewBufferPool(size int) Pool {
	return &sync.Pool{
		New: func() interface{} {
			return make([]byte, size)
		},
	}
}

// defaultWriteBufPool is a default pool for write buffers.
var defaultWriteBufPool = NewBufferPool(defaultWriteBufferSize + maxFrameHeaderSize)

// NewBufReaderPool creates a new Pool of *bufio.Reader.
func NewBufReaderPool(size int) Pool {
	return &sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, size)
		},
	}
}

// defaultBufReaderPool is a default pool for bufio.Readers.
var defaultBufReaderPool = NewBufReaderPool(defaultReadBufferSize)

// putFunc is a callback used to return buffers
type putFunc func(interface{})

// put puts the value if the putFunc is non-nil.
func (p putFunc) put(i interface{}) {
	if p != nil {
		p(i)
	}
}

// ioBuf is a set of I/O buffers and mechanisms to reuse them.
type ioBuf struct {
	// br is the buffered reader used to read the message stream.
	br *bufio.Reader

	// writeBuf is a write buffer used to construct messages.
	writeBuf []byte

	// putBR and putWBuf are putFuncs used to recycle the buffers.
	putBR, putWBuf putFunc
}

// getIOBuf gets a set of I/O buffers, pooling based on the Upgrader settings.
func (u *Upgrader) getIOBuf() ioBuf {
	var writeBuf []byte
	var writeBufPut putFunc
	switch {
	case u.WriteBufferPool != nil:
		writeBuf, writeBufPut = u.WriteBufferPool.Get().([]byte), u.WriteBufferPool.Put
	case u.WriteBufferSize != 0 && u.WriteBufferSize != defaultWriteBufferSize:
		writeBuf = make([]byte, u.WriteBufferSize+maxFrameHeaderSize)
	default:
		writeBuf, writeBufPut = defaultWriteBufPool.Get().([]byte), defaultWriteBufPool.Put
	}

	var br *bufio.Reader
	var brPut putFunc
	switch {
	case u.BufReaderPool != nil:
		br, brPut = u.BufReaderPool.Get().(*bufio.Reader), u.BufReaderPool.Put
	case u.ReadBufferSize != 0 && u.ReadBufferSize != defaultReadBufferSize:
		br = bufio.NewReaderSize(nil, u.ReadBufferSize)
	default:
		br, brPut = defaultBufReaderPool.Get().(*bufio.Reader), defaultBufReaderPool.Put
	}

	return ioBuf{
		br:       br,
		writeBuf: writeBuf,
		putBR:    brPut,
		putWBuf:  writeBufPut,
	}
}

// getIOBuf gets a set of I/O buffers, pooling based on the Dialer settings.
func (d *Dialer) getIOBuf() ioBuf {
	var writeBuf []byte
	var writeBufPut putFunc
	switch {
	case d.WriteBufferPool != nil:
		writeBuf, writeBufPut = d.WriteBufferPool.Get().([]byte), d.WriteBufferPool.Put
	case d.WriteBufferSize != 0 && d.WriteBufferSize != defaultWriteBufferSize:
		writeBuf = make([]byte, d.WriteBufferSize+maxFrameHeaderSize)
	default:
		writeBuf, writeBufPut = defaultWriteBufPool.Get().([]byte), defaultWriteBufPool.Put
	}

	var br *bufio.Reader
	var brPut putFunc
	switch {
	case d.BufReaderPool != nil:
		br, brPut = d.BufReaderPool.Get().(*bufio.Reader), d.BufReaderPool.Put
	case d.ReadBufferSize != 0 && d.ReadBufferSize != defaultReadBufferSize:
		br = bufio.NewReaderSize(nil, d.ReadBufferSize)
	default:
		br, brPut = defaultBufReaderPool.Get().(*bufio.Reader), defaultBufReaderPool.Put
	}

	return ioBuf{
		br:       br,
		writeBuf: writeBuf,
		putBR:    brPut,
		putWBuf:  writeBufPut,
	}
}

// cleanup recycles the I/O buffers and invalidates the ioBuf.
func (iob *ioBuf) cleanup() {
	// reset bufio.Reader to allow underlying conn to be garbage collected
	if iob.br != nil {
		iob.br.Reset(nil)
	}

	// recycle bufio reader and write buffer
	iob.putBR.put(iob.br)
	iob.putWBuf.put(iob.writeBuf)

	// clear ioBuf to prevent reuse
	*iob = ioBuf{}
}

// chanPool is a channel-based pool implementation for testing purposes
type chanPool chan interface{}

func (c chanPool) Put(i interface{}) {
	select {
	case c <- i:
	default:
	}
}

func (c chanPool) Get() interface{} {
	select {
	case i := <-c:
		return i
	default:
		panic(errors.New("no value"))
	}
}
