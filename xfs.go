package xfs

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"
	"unsafe"
)

//BulkReq keeps track of in process request
type BulkReq struct {
	last  uint64
	batch int32
	fsfd  *os.File
}

//NewBulkReq creates a new XFS bulk request
func NewBulkReq(path string, opts ...Option) (*BulkReq, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	b := &BulkReq{
		//default batch size of 4096
		batch: 4096,
		fsfd:  f,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

//Option sets various options for NewBulkReq
type Option func(*BulkReq)

//WithStartNum sets the starting inode for bulk request
func WithStartNum(start uint64) Option {
	return func(m *BulkReq) { m.last = start }
}

//WithBatchSize sets the batch size for bulk request
func WithBatchSize(batch int32) Option {
	return func(m *BulkReq) { m.batch = batch }
}

//Next gets the next batch of Bstats
func (b *BulkReq) Next() ([]Bstat, error) {
	buf := make([]byte, BstatSize*b.batch)
	var count int32
	f := &fsopBulkreq{
		lastip:  unsafe.Pointer(&b.last),
		icount:  b.batch,
		ubuffer: unsafe.Pointer(&buf[0]),
		ocount:  unsafe.Pointer(&count),
	}

	err := xfsctl(b.fsfd.Fd(), IOCFSBULKSTAT, uintptr(unsafe.Pointer(f)))
	if err != nil {
		return []Bstat{}, err
	}

	rbuf := bytes.NewReader(buf)
	var bstats []Bstat
	for i := 0; i < int(count); i++ {
		var b Bstat
		err := binary.Read(rbuf, binary.LittleEndian, &b)
		if err != nil {
			return []Bstat{}, err
		}
		bstats = append(bstats, b)
	}
	return bstats, nil
}

//Release BulkReq handle and cleanup
func (b *BulkReq) Release() {
	b.fsfd.Close()
}

func xfsctl(fd, cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if err != 0 {
		return err
	}
	return nil
}
