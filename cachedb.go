package ip2location

import (
	"errors"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

// const (
// 	numPartitions = 16
// )

type CacheDB struct {
	r      io.ReaderAt
	kind   EntryKind
	fields Fields
	date   time.Time
	values stringCache
	v4     dbEntryCache
	v6     dbEntryCache
}

func OpenCacheDB(path string, fields ...Field) (*CacheDB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	db := new(CacheDB)
	if err := db.Reset(f); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *CacheDB) Reset(r io.ReaderAt, fields ...Field) (err error) {
	kind, date, err := readDBMeta(r)
	if err != nil {
		return
	}
	if len(fields) == 0 {
		fields = dbFields[kind]
	}

	fields, err = requireFields(kind, fields)
	if err != nil {
		return
	}
	*db = CacheDB{
		r:      r,
		kind:   kind,
		date:   date,
		fields: fields,
	}
	err = db.v4.Reset(r, 4, kind)
	if err != nil {
		return
	}
	err = db.v6.Reset(r, 6, kind)
	if err != nil {
		return
	}

	return
}

type dbEntryCache struct {
	rows    []byte
	size    int
	n       int
	mu      sync.RWMutex
	index   map[int]int
	entries []*Entry
}

func (db *CacheDB) Lookup(ip net.IP) (*Entry, error) {
	var c *dbEntryCache
	switch ip = NormalizeIP(ip); len(ip) {
	case net.IPv4len:
		c = &db.v4
	case net.IPv6len:
		c = &db.v6
	default:
		return nil, errors.New("Invalid ip address")
	}
	i := c.indexOf(ip)
	if i == -1 {
		return nil, nil
	}
	if e := c.get(i); e != nil {
		return e, nil
	}
	e := Entry{}
	row := c.rowAt(i)
	row = row[len(ip):]
	if err := db.loadEntry(&e, row); err != nil {
		return nil, err
	}
	c.set(i, &e)
	return &e, nil
}

func (db *CacheDB) loadEntry(e *Entry, buf []byte) error {
	var n uint32
	for _, f := range db.fields {
		if len(buf) >= 4 {
			n, buf = u32LE(buf), buf[4:]
			if err := db.readField(e, f, n); err != nil {
				return err
			}
		} else {
			return io.ErrShortBuffer
		}
	}
	return nil

}
func (db *CacheDB) readField(e *Entry, field Field, n uint32) (err error) {
	switch field {
	case FieldCountry:
		e.Country, err = db.readString(n)
	case FieldRegion:
		e.Region, err = db.readString(n)
	case FieldCity:
		e.City, err = db.readString(n)
	case FieldISP:
		e.ISP, err = db.readString(n)
	case FieldLatitude:
		e.Latitude = math.Float32frombits(n)
	case FieldLongitude:
		e.Longitude = math.Float32frombits(n)
	case FieldDomain:
		e.Domain, err = db.readString(n)
	case FieldZipCode:
		e.ZipCode, err = db.readString(n)
	case FieldTimeZone:
		e.TimeZone, err = db.readString(n)
	case FieldNetSpeed:
		e.NetSpeed, err = db.readString(n)
	case FieldIDDCode:
		e.IDDCode, err = db.readString(n)
	case FieldAreaCode:
		e.AreaCode, err = db.readString(n)
	case FieldWeatherCode:
		e.WeatherStationCode, err = db.readString(n)
	case FieldWeatherName:
		e.WeatherStationName, err = db.readString(n)
	case FieldMCC:
		e.MCC, err = db.readString(n)
	case FieldMNC:
		e.MCC, err = db.readString(n)
	case FieldMobileBrand:
		e.MobileBrand, err = db.readString(n)
	case FieldElevation:
		e.Elevation = math.Float32frombits(n)
	case FieldUsageType:
		e.UsageType, err = db.readString(n)
	}
	return

}
func (db *CacheDB) readString(id uint32) (s string, err error) {
	s, ok := db.values.get(id)
	if ok {
		return
	}
	pos := int64(id)
	size, err := readByte(db.r, pos)
	if err != nil {
		return
	}
	buf := make([]byte, size)
	_, err = db.r.ReadAt(buf[:], pos+1)
	if err != nil {
		return
	}
	s = string(buf)
	db.values.set(id, s)
	return
}
func (db *dbEntryCache) rowAt(i int) []byte {
	size := len(db.rows) / db.n

	i *= size
	if 0 <= i && i < len(db.rows) {
		row := db.rows[i:]
		if 0 <= size && size <= len(row) {
			return row[:size]
		}
	}
	return nil
}

func (db *dbEntryCache) indexOf(ip net.IP) int {
	size := len(db.rows) / db.n
	lo, hi := 0, len(db.rows)/size
	for lo <= hi {
		mid := (lo + hi) / 2
		if compareAt(ip, db.rows, mid) < 0 {
			hi = mid - 1
		} else if lo = mid + 1; compareAt(ip, db.rows, lo) < 0 {
			return mid
		}
	}
	return -1
}

// func NewCacheDB(path string) (*CacheDB, error) {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db := new(CacheDB)
// 	if err := db.Reset(f); err != nil {
// 		return nil, err
// 	}

// 	return db, nil
// }

// func (db *CacheDB) Reset(r io.ReaderAt) error {
// 	kind, date, err := readDBMeta(r)
// 	if err != nil {
// 		return err
// 	}
// 	db.Close()
// 	db.r, db.kind, db.date = r, kind, date

// 	return nil
// }

// func (db *CacheDB) Close() error {
// 	r := db.r
// 	db.r = nil
// 	if c, ok := r.(io.Closer); ok {
// 		return c.Close()
// 	}
// 	return nil
// }

// type entryLookup struct {
// 	index  ipLookup
// 	offset int64
// }

// func (db *entryLookup) read(r io.ReaderAt, v int, kind EntryKind) error {
// 	uOffset, uCount, err := readEntries(r, v)
// 	if err != nil {
// 		return err
// 	}
// 	if uCount == 0 {
// 		return nil
// 	}
// 	offset := int64(uOffset)
// 	db.offset = offset
// 	size := int64(ipSize(v))
// 	fields := dbFields[kind]
// 	row := make([]byte, int(size)+len(fields)*4)
// 	ip := row[:size]
// 	n := int64(uCount)
// 	rowSize := int64(len(row))
// 	db.index = newIPLookup(v, int(n))
// 	for i := int64(0); i < n; i++ {
// 		pos := offset + i*rowSize
// 		if _, err := r.ReadAt(row, pos); err != nil {
// 			return err
// 		}
// 		fixIP(row, v)
// 		copy(db.index.ips[i*size:], ip)
// 	}

// }

func (c *dbEntryCache) get(i int) *Entry {
	c.mu.RLock()
	if i, ok := c.index[i]; ok {
		e := c.entries[i]
		c.mu.RUnlock()
		return e
	}
	c.mu.RUnlock()
	return nil
}

func (c *dbEntryCache) set(i int, e *Entry) {
	c.mu.Lock()
	c.index[i] = len(c.entries)
	c.entries = append(c.entries, e)
	c.mu.Unlock()
}

type stringCache struct {
	mu    sync.RWMutex
	index map[uint32]string
}

func (c *stringCache) get(id uint32) (s string, ok bool) {
	c.mu.RLock()
	s, ok = c.index[id]
	c.mu.RUnlock()
	return
}
func (c *stringCache) set(id uint32, s string) {
	c.mu.Lock()
	if c.index == nil {
		c.index = make(map[uint32]string)
	}
	c.index[id] = s
	c.mu.Unlock()
}

// func (idx *ipIndex) Read(r io.ReaderAt, v int, k EntryKind) (err error) {
// 	var count, offset uint32
// 	switch v {
// 	case 4:
// 		count, err = readUint32(r, 5)
// 		if err != nil {
// 			return
// 		}
// 		idx.size = 4
// 		offset, err = readUint32(r, 9)
// 	case 6:
// 		count, err = readUint32(r, 13)
// 		if err != nil {
// 			return
// 		}
// 		idx.size = 6
// 		offset, err = readUint32(r, 17)
// 	default:
// 		err = errors.New("Invalid ip version")
// 	}
// 	if err != nil {
// 		return
// 	}
// 	if count == 0 {
// 		return
// 	}
// 	var rowSize = uint32(idx.size + 4*len(dbFields[k]))
// 	data := make([]byte, rowSize*count)
// 	_, err = r.ReadAt(data, int64(offset))
// 	if err != nil {
// 		return
// 	}
// 	idx.ips = make([]byte, uint32(idx.size)*count)
// 	for i := uint32(0); i < count; i++ {
// 		if uint32(len(data)) >= rowSize {
// 			copy(idx.ips[i*idx.size:], data)
// 			data = data[rowSize:]
// 		}

// 	}
// 	return
// }

func (db *dbEntryCache) Reset(r io.ReaderAt, v int, k EntryKind) (err error) {
	var count, offset uint32
	switch v {
	case 4:
		count, err = readUint32(r, 5)
		if err != nil {
			return
		}
		offset, err = readUint32(r, 9)
	case 6:
		count, err = readUint32(r, 13)
		if err != nil {
			return
		}
		offset, err = readUint32(r, 17)
	default:
		err = errors.New("Invalid ip version")
	}
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	size := rowSize(v, k)
	rows := make([]byte, count*uint32(size))
	_, err = r.ReadAt(rows, int64(offset)-1)
	if err != nil {
		return
	}
	for i := 0; i < len(rows); i += size {
		fixIP(rows[i:], v)
	}
	db.mu.Lock()
	db.rows = rows
	db.size = size
	db.n = int(count)
	db.index = make(map[int]int)
	db.entries = db.entries[:0]
	db.mu.Unlock()
	return
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
