package websocket

import (
	"bufio"
	"testing"
)

func TestDialerGetIOBuf(t *testing.T) {
	// prepare objects
	wpool := make(chanPool, 1)
	wbuf := make([]byte, defaultWriteBufferSize/4)
	rpool := make(chanPool, 1)
	br := bufio.NewReaderSize(nil, defaultReadBufferSize/4)

	// save buffers to pool
	wpool.Put(wbuf)
	rpool.Put(br)

	// get ioBuf using specific pools
	chanPoolDialer := &Dialer{
		WriteBufferPool: wpool,
		BufReaderPool:   rpool,
	}
	iob := chanPoolDialer.getIOBuf()
	if iob.br != br {
		t.Errorf("Expected %T %p but got %p", br, br, iob.br)
	}
	if &iob.writeBuf[0] != &wbuf[0] {
		t.Errorf("Expected %T %p but got %p", wbuf, wbuf, iob.writeBuf)
	}
	iob.cleanup()
	if len(wpool) != 1 {
		t.Error("Expected write buffer to be recycled, buffer not recycled")
	}
	if len(rpool) != 1 {
		t.Error("Expected *bufio.Reader to be recycled, reader not recycled")
	}

	// get ioBuf using size spec
	sizeDialer := &Dialer{
		WriteBufferSize: 2 * defaultWriteBufferSize,
		ReadBufferSize:  2 * defaultReadBufferSize,
	}
	iob = sizeDialer.getIOBuf()
	if len(iob.writeBuf) != sizeDialer.WriteBufferSize+maxFrameHeaderSize {
		t.Errorf("Expected write buffer with len %d but got len %d", sizeDialer.WriteBufferSize+maxFrameHeaderSize, len(iob.writeBuf))
	}
	if iob.br.Size() != sizeDialer.ReadBufferSize {
		t.Errorf("Expected buffered reader with size %d but got %d", sizeDialer.ReadBufferSize, iob.br.Size())
	}

	// get ioBuf using default dialer
	iob = ((*Dialer)(nil)).getIOBuf()
	if len(iob.writeBuf) != defaultWriteBufferSize+maxFrameHeaderSize {
		t.Errorf("Expected write buffer with len %d but got len %d", defaultWriteBufferSize+maxFrameHeaderSize, len(iob.writeBuf))
	}
	if iob.br.Size() != defaultReadBufferSize {
		t.Errorf("Expected buffered reader with size %d but got %d", defaultReadBufferSize, iob.br.Size())
	}
}

func TestUpgraderGetIOBuf(t *testing.T) {
	// prepare objects
	wpool := make(chanPool, 1)
	wbuf := make([]byte, defaultWriteBufferSize/4)
	rpool := make(chanPool, 1)
	br := bufio.NewReaderSize(nil, defaultReadBufferSize/4)

	// save buffers to pool
	wpool.Put(wbuf)
	rpool.Put(br)

	// get ioBuf using specific pools
	chanPoolUpgrader := &Upgrader{
		WriteBufferPool: wpool,
		BufReaderPool:   rpool,
	}
	iob := chanPoolUpgrader.getIOBuf()
	if iob.br != br {
		t.Errorf("Expected %T %p but got %p", br, br, iob.br)
	}
	if &iob.writeBuf[0] != &wbuf[0] {
		t.Errorf("Expected %T %p but got %p", wbuf, wbuf, iob.writeBuf)
	}
	iob.cleanup()
	if len(wpool) != 1 {
		t.Error("Expected write buffer to be recycled, buffer not recycled")
	}
	if len(rpool) != 1 {
		t.Error("Expected *bufio.Reader to be recycled, reader not recycled")
	}

	// get ioBuf using size spec
	sizeUpgrader := &Upgrader{
		WriteBufferSize: 2 * defaultWriteBufferSize,
		ReadBufferSize:  2 * defaultReadBufferSize,
	}
	iob = sizeUpgrader.getIOBuf()
	if len(iob.writeBuf) != sizeUpgrader.WriteBufferSize+maxFrameHeaderSize {
		t.Errorf("Expected write buffer with len %d but got len %d", sizeUpgrader.WriteBufferSize+maxFrameHeaderSize, len(iob.writeBuf))
	}
	if iob.br.Size() != sizeUpgrader.ReadBufferSize {
		t.Errorf("Expected buffered reader with size %d but got %d", sizeUpgrader.ReadBufferSize, iob.br.Size())
	}

	// get ioBuf using default upgrader
	iob = (&Upgrader{}).getIOBuf()
	if len(iob.writeBuf) != defaultWriteBufferSize+maxFrameHeaderSize {
		t.Errorf("Expected write buffer with len %d but got len %d", defaultWriteBufferSize+maxFrameHeaderSize, len(iob.writeBuf))
	}
	if iob.br.Size() != defaultReadBufferSize {
		t.Errorf("Expected buffered reader with size %d but got %d", defaultReadBufferSize, iob.br.Size())
	}
}
