package ip2location

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

type rowIterator struct {
	r   io.ReaderAt
	row []byte
	pos int64
	n   int64
	i   int64
	err error
}

func rowSize(v int, kind EntryKind) int {
	fields := dbFields[kind]
	switch v {
	case 4:
		return net.IPv4len + 4*len(fields)
	case 6:
		return net.IPv6len + 4*len(fields)
	default:
		return 0
	}
}

func newRowIterator(r io.ReaderAt, v int, kind EntryKind) (iter *rowIterator, err error) {
	var n uint64
	switch v {
	case 4:
		n, err = readUint64(r, 5)
	case 6:
		n, err = readUint64(r, 13)
	default:
		return
	}
	if err != nil {
		return
	}
	count, offset := uint32(n), uint32(n>>32)
	if count == 0 {
		return
	}
	iter = &rowIterator{
		r:   r,
		row: make([]byte, rowSize(v, kind)),
		pos: int64(offset) - 1,
		n:   int64(count),
	}
	return
}

func (rows *rowIterator) Next() bool {
	if rows.i < rows.n && rows.err == nil {
		pos := rows.pos + rows.i*int64(len(rows.row))
		rows.i++
		_, rows.err = rows.r.ReadAt(rows.row, pos)
		return rows.err == nil
	}
	return false
}

func (rows *rowIterator) Row() []byte {
	return rows.row
}

func (rows *rowIterator) Err() error {
	return rows.err
}

func readUint64(r io.ReaderAt, pos int64) (uint64, error) {
	buf := [8]byte{}
	_, err := r.ReadAt(buf[:], pos)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func u32LE(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func readUint32(r io.ReaderAt, pos int64) (uint32, error) {
	tmp := [4]byte{}
	_, err := r.ReadAt(tmp[:], pos)
	if err != nil {
		return 0, err
	}
	return u32LE(tmp[:]), nil
}
func readByte(r io.ReaderAt, pos int64) (byte, error) {
	buf := [1]byte{}
	_, err := r.ReadAt(buf[:], pos)
	return buf[0], err
}

func rowIP(row []byte, v int) (ip []byte, data []byte) {
	switch v {
	case 4:
		if len(row) >= net.IPv4len {
			ip, data = row[:net.IPv4len], row[net.IPv4len:]
			ip[0], ip[1], ip[2], ip[3] = ip[3], ip[2], ip[1], ip[0]
		}
	case 6:
		if len(row) >= net.IPv6len {
			ip, data = row[:net.IPv6len], row[net.IPv6len:]
			ip[0], ip[1], ip[2], ip[3], ip[4], ip[5], ip[6], ip[7], ip[8], ip[9], ip[10], ip[11], ip[12], ip[13], ip[14], ip[15] = ip[15], ip[14], ip[13], ip[12], ip[11], ip[10], ip[9], ip[8], ip[7], ip[6], ip[5], ip[4], ip[3], ip[2], ip[1], ip[0]
		}
	}
	return

}

func readDBMeta(r io.ReaderAt) (kind EntryKind, tm time.Time, err error) {
	buf := [5]byte{}
	_, err = r.ReadAt(buf[:], 0)
	if err != nil {
		return
	}
	kind = EntryKind(buf[0])
	if _, ok := dbFields[kind]; !ok {
		err = errors.New("Unknown db type")
		return
	}
	tm = time.Date(int(buf[2]), time.Month(buf[3]), int(buf[4]), 0, 0, 0, 0, time.UTC)
	return

}
