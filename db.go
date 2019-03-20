package ip2location

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

type DB struct {
	r  dbReader
	v4 dbEntries
	v6 dbEntries
}

func Open(path string) (*DB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	db := new(DB)
	if err := db.Reset(f); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	r := db.r.ReaderAt
	if c, ok := r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (db *DB) Reset(r io.ReaderAt) error {
	if err := db.r.Reset(r); err != nil {
		return err
	}
	if err := db.readEntries(); err != nil {
		return err
	}
	return nil
}

func (db *DB) readEntries() error {
	errc := make(chan error, 2)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	read := func(entries *dbEntries, v int) {
		defer wg.Done()
		errc <- db.r.readEntries(entries, v)
	}
	go read(&db.v4, 4)
	go read(&db.v6, 6)
	wg.Wait()
	close(errc)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Kind() EntryKind {
	return db.r.kind
}

func (db *DB) Date() time.Time {
	return db.r.date
}

func (db *DB) lookup(ip net.IP) []byte {
	switch ip = NormalizeIP(ip); len(ip) {
	case net.IPv4len:
		return db.v4.lookup(ip)
	case net.IPv6len:
		return db.v6.lookup(ip)
	default:
		return nil
	}
}
func (db *DB) Query(e *Entry, ip net.IP, fields ...Field) error {
	row := db.lookup(ip)
	if row == nil {
		return errors.New("Not found")
	}
	mask := Fields(fields).Mask()
	if mask == 0 {
		mask = allFields
	}
	return db.r.ReadEntry(e, row, mask)
	return nil
}

func (db *DB) Each(v int, fields Fields, fn func(e *Entry, lo, hi net.IP)) error {
	e := Entry{}
	m := fields.Mask()
	if len(fields) == 0 {
		m = allFields
	}
	var entries *dbEntries
	var offset int
	switch v {
	case 4:
		entries = &db.v4
		offset = net.IPv4len
	case 6:
		entries = &db.v6
		offset = net.IPv6len
	default:
		return errors.New("Invalid ip version")
	}
	if m == 0 {
		m = allFields
	}

	for i := 0; i < entries.n; i++ {
		row := entries.rowAt(i)
		lo, data := row[:offset], row[offset:]
		hi := entries.rowAt(i + 1)
		if len(hi) >= offset {
			hi = hi[:offset]
		}
		if err := db.r.ReadEntry(&e, data, m); err != nil {
			return err
		}
		fn(&e, lo, hi)
	}
	return nil
}

type dbReader struct {
	io.ReaderAt
	kind   EntryKind
	fields []Field
	date   time.Time
	mu     sync.RWMutex
	values map[uint32]string
}

func (db *dbReader) Reset(r io.ReaderAt) (err error) {
	kind, date, err := readDBMeta(r)
	if err != nil {
		return
	}
	*db = dbReader{
		ReaderAt: r,
		kind:     kind,
		fields:   dbFields[kind],
		date:     date,
	}
	return
}

func (r *dbReader) readEntries(entries *dbEntries, v int) (err error) {
	var count, offset uint32
	switch v {
	case 4:
		count, offset = 5, 9
	case 6:
		count, offset = 13, 17
	default:
		err = errors.New("Invalid ip version")
		return
	}
	count, err = readUint32(r, int64(count))
	if err != nil {
		return
	}
	if count == 0 {
		*entries = dbEntries{}
		return
	}
	offset, err = readUint32(r, int64(offset))
	if err != nil {
		return
	}
	size := rowSize(v, r.kind)
	rows := make([]byte, count*uint32(size))
	_, err = r.ReadAt(rows, int64(offset)-1)
	if err != nil {
		return
	}
	for i := 0; i < len(rows); i += size {
		fixIP(rows[i:], v)
	}
	*entries = dbEntries{
		rows: rows,
		size: size,
		n:    int(count),
	}
	return nil
}

func (db *dbReader) ReadEntry(e *Entry, buf []byte, fields FieldMask) (err error) {
	var n uint32
	for _, f := range db.fields {
		if len(buf) >= 4 {
			n, buf = u32LE(buf), buf[4:]
			if fields.Has(f) {
				switch f {
				case FieldCountry:
					e.Country, err = db.ReadValue(f, n)
				case FieldRegion:
					e.Region, err = db.ReadValue(f, n)
				case FieldCity:
					e.City, err = db.ReadValue(f, n)
				case FieldISP:
					e.ISP, err = db.ReadValue(f, n)
				case FieldLatitude:
					e.Latitude = math.Float32frombits(n)
				case FieldLongitude:
					e.Longitude = math.Float32frombits(n)
				case FieldDomain:
					e.Domain, err = db.ReadValue(f, n)
				case FieldZipCode:
					e.ZipCode, err = db.ReadValue(f, n)
				case FieldTimeZone:
					e.TimeZone, err = db.ReadValue(f, n)
				case FieldNetSpeed:
					e.NetSpeed, err = db.ReadValue(f, n)
				case FieldIDDCode:
					e.IDDCode, err = db.ReadValue(f, n)
				case FieldAreaCode:
					e.AreaCode, err = db.ReadValue(f, n)
				case FieldWeatherCode:
					e.WeatherStationCode, err = db.ReadValue(f, n)
				case FieldWeatherName:
					e.WeatherStationName, err = db.ReadValue(f, n)
				case FieldMCC:
					e.MCC, err = db.ReadValue(f, n)
				case FieldMNC:
					e.MCC, err = db.ReadValue(f, n)
				case FieldMobileBrand:
					e.MobileBrand, err = db.ReadValue(f, n)
				case FieldElevation:
					e.Elevation = math.Float32frombits(n)
				case FieldUsageType:
					e.UsageType, err = db.ReadValue(f, n)
				default:
					panic("Invalid field")
				}
				if err != nil {
					return
				}
			}
		} else {
			return io.ErrShortBuffer
		}
	}
	return nil
}

func (db *dbReader) Get(id uint32) (s string, ok bool) {
	db.mu.RLock()
	s, ok = db.values[id]
	db.mu.RUnlock()
	return
}
func (db *dbReader) Set(id uint32, s string) {
	db.mu.Lock()
	if db.values == nil {
		db.values = make(map[uint32]string)
	}
	db.values[id] = s
	db.mu.Unlock()
}

func (db *dbReader) ReadValue(f Field, id uint32) (s string, err error) {
	s, ok := db.Get(id)
	if ok {
		return
	}
	pos := int64(id)
	size, err := readByte(db.ReaderAt, pos)
	if err != nil {
		return
	}
	buf := make([]byte, size)
	_, err = db.ReaderAt.ReadAt(buf[:], pos+1)
	if err != nil {
		return
	}
	s = string(buf)
	db.Set(id, s)
	return
}

type dbEntries struct {
	rows []byte
	size int
	n    int
}

func (db *dbEntries) rowAt(i int) []byte {
	size := db.size

	i *= size
	if 0 <= i && i < len(db.rows) {
		row := db.rows[i:]
		if 0 <= size && size <= len(row) {
			return row[:size]
		}
	}
	return nil
}

func compareAt(a, b []byte, i int) int {
	i *= len(a)
	if 0 <= i && i < len(b) {
		b = b[i:]
		if len(b) >= len(a) {
			return bytes.Compare(a, b[:len(a)])
		}
	}
	return 1
}

func (db *dbEntries) lookup(ip net.IP) []byte {
	lo, hi := 0, db.n
	for lo <= hi {
		mid := (lo + hi) / 2
		if compareAt(ip, db.rows, mid) < 0 {
			hi = mid - 1
		} else if lo = mid + 1; compareAt(ip, db.rows, lo) < 0 {
			return db.rows[mid*db.size+len(ip) : lo*db.size]
		}
	}
	return nil
}

func fixIP(ip []byte, v int) {
	switch v {
	case 4:
		if len(ip) >= net.IPv4len {
			ip[0], ip[1], ip[2], ip[3] = ip[3], ip[2], ip[1], ip[0]
		}
	case 6:
		if len(ip) >= net.IPv6len {
			ip[0], ip[1], ip[2], ip[3], ip[4], ip[5], ip[6], ip[7], ip[8], ip[9], ip[10], ip[11], ip[12], ip[13], ip[14], ip[15] = ip[15], ip[14], ip[13], ip[12], ip[11], ip[10], ip[9], ip[8], ip[7], ip[6], ip[5], ip[4], ip[3], ip[2], ip[1], ip[0]
		}
	}
	return
}

func NormalizeIP(ip net.IP) net.IP {
	if len(ip) == net.IPv4len {
		return ip
	}
	const zeroPrefix = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xFF\xFF"
	if len(ip) == net.IPv6len && string(ip[:12]) == zeroPrefix {
		return ip[12:16]
	}
	return ip
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

type key int

var dbKey key

func FromContext(ctx context.Context) *DB {
	if x, ok := ctx.Value(dbKey).(*DB); ok {
		return x
	}
	return nil
}

func NewContext(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}
