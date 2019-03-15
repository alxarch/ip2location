package ip2location

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"net"
	"time"
)

type DB struct {
	kind EntryKind
	date time.Time
	v4   *dbEntries
	v6   *dbEntries
}

func Open(path string) (*DB, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := dbReader{
		ReaderAt: bytes.NewReader(data),
	}
	kind, date, err := r.readMeta()
	if err != nil {
		return nil, err
	}
	fields := dbFields[kind]
	v4, err := r.readEntries(4, fields)
	if err != nil {
		return nil, err
	}
	v6, err := r.readEntries(6, fields)
	if err != nil {
		return nil, err
	}
	return &DB{
		kind: kind,
		date: date,
		v4:   v4,
		v6:   v6,
	}, nil
}

func (db *DB) Kind() EntryKind {
	return db.kind
}

func (db *DB) Date() time.Time {
	return db.date
}

func (db *DB) LookupString(ip string) *Entry {
	return db.Lookup(net.ParseIP(ip))
}

func (db *DB) Lookup(ip net.IP) *Entry {
	switch ip = NormalizeIP(ip); len(ip) {
	case net.IPv4len:
		return db.v4.Lookup(ip)
	case net.IPv6len:
		return db.v6.Lookup(ip)
	}
	return nil
}

type dbEntries struct {
	entries []Entry
	ips     []byte
	index   [256]dbOffset
}

type dbOffset struct {
	start, end uint32
}

func (db *dbEntries) buildIndex() {
	var (
		size   = db.ipSize()
		last   byte
		i, pos int
	)
	for ; 0 <= i && i < len(db.ips); i += size {
		for c := db.ips[i]; last < c; last++ {
			idx := &db.index[last]
			idx.start, idx.end = uint32(pos), uint32(i/size)
			pos = i / size
		}
	}
	for ; last < 255; last++ {
		idx := &db.index[last]
		idx.start, idx.end = uint32(pos), uint32(len(db.entries))
		pos = len(db.entries)
	}
}

func (db *dbEntries) ipSize() int {
	return len(db.ips) / len(db.entries)
}
func (db *dbEntries) entry(i uint32) *Entry {
	if i <= uint32(len(db.entries)) {
		return &db.entries[i]
	}
	return nil
}

func (db *dbEntries) cmp(i uint32, ip net.IP) int {
	if i <= uint32(len(db.ips)) {
		b := db.ips[i:]
		if len(b) >= len(ip) {
			return bytes.Compare(ip, b[:len(ip)])
		}
	}
	return -1
}

func (db *dbEntries) Lookup(ip net.IP) *Entry {
	if db == nil {
		return nil
	}
	lo, hi := db.ipRange(ip)
	for lo <= hi {
		mid := (lo + hi) / 2
		if db.cmp(mid, ip) < 0 {
			hi = mid - 1
		} else if lo = mid + 1; db.cmp(lo, ip) < 0 {
			return db.entry(mid)
		}
	}

	return nil

}
func (db *dbEntries) ip(i uint32, size int) net.IP {
	if i <= uint32(len(db.ips)) {
		b := db.ips[i:]
		if len(b) >= size {
			return b[:size]
		}
	}
	return nil
}
func (db *dbEntries) ipRange(ip net.IP) (lo, hi uint32) {
	if len(ip) > 0 {
		k := ip[0]
		idx := &db.index[k]
		lo, hi = idx.start, idx.end
	}
	// hi = uint32(len(db.entries))
	return

}

type dbReader struct {
	io.ReaderAt
	cache map[uint32]string
}

func (r *dbReader) readMeta() (kind EntryKind, tm time.Time, err error) {
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

func (r *dbReader) readEntries(ipVersion int, fields []field) (*dbEntries, error) {
	var offset, count, size uint32
	switch ipVersion {
	case 4:
		count, offset, size = 5, 9, net.IPv4len
	case 6:
		count, offset, size = 13, 17, net.IPv6len
	default:
		return nil, errors.New("Invalid ip version")
	}
	var err error
	if count, err = readUint32(r, int64(count)); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	if offset, err = readUint32(r, int64(offset)); err != nil {
		return nil, err
	}
	rowSize := size + uint32(len(fields)*4)
	buf := make([]byte, rowSize)
	ip, data := buf[:size], buf[size:]
	// The offset is not 0 based
	offset--

	db := dbEntries{
		ips:     make([]byte, count*size),
		entries: make([]Entry, count),
	}
	for i := uint32(0); i < count; i++ {
		pos := int64(offset + i*rowSize)
		_, err := r.ReadAt(buf, pos)
		if err != nil {
			return nil, err
		}
		fixIP(buf, ipVersion)
		copy(db.ips[i*size:], ip)
		e := &db.entries[i]
		if err := r.readEntry(e, data, fields); err != nil {
			return nil, err
		}
	}
	db.buildIndex()
	return &db, nil
}

func (r *dbReader) readField(e *Entry, f field, n uint32) (err error) {
	switch f {
	case fieldCountry:
		e.Country, err = r.readString(n)
	case fieldRegion:
		e.Region, err = r.readString(n)
	case fieldCity:
		e.City, err = r.readString(n)
	case fieldISP:
		e.ISP, err = r.readString(n)
	case fieldLatitude:
		e.Latitude = math.Float32frombits(n)
	case fieldLongitude:
		e.Longitude = math.Float32frombits(n)
	case fieldDomain:
		e.Domain, err = r.readString(n)
	case fieldZipCode:
		e.ZipCode, err = r.readString(n)
	case fieldTimeZone:
		e.TimeZone, err = r.readString(n)
	case fieldNetSpeed:
		e.NetSpeed, err = r.readString(n)
	case fieldIDDCode:
		e.IDDCode, err = r.readString(n)
	case fieldAreaCode:
		e.AreaCode, err = r.readString(n)
	case fieldWeatherCode:
		e.WeatherStationCode, err = r.readString(n)
	case fieldWeatherName:
		e.WeatherStationName, err = r.readString(n)
	case fieldMCC:
		e.MCC, err = r.readString(n)
	case fieldMNC:
		e.MCC, err = r.readString(n)
	case fieldMobileBrand:
		e.MobileBrand, err = r.readString(n)
	case fieldElevation:
		e.Elevation = math.Float32frombits(n)
	case fieldUsageType:
		e.UsageType, err = r.readString(n)
	}
	return

}
func (r *dbReader) readEntry(e *Entry, buf []byte, fields []field) error {
	var n uint32
	if len(fields) != len(buf)/4 {
		panic("Invalid fields")
	}
	for _, f := range fields {
		if len(buf) >= 4 {
			n, buf = u32LE(buf), buf[4:]
			if err := r.readField(e, f, n); err != nil {
				return err
			}
		} else {
			return io.ErrShortBuffer
		}
	}
	return nil

}

func (r *dbReader) readString(id uint32) (s string, err error) {
	s, ok := r.cache[id]
	if ok {
		return
	}
	pos := int64(id)
	size, err := r.readByte(pos)
	if err != nil {
		return
	}
	buf := make([]byte, size)
	_, err = r.ReadAt(buf[:], pos+1)
	if err != nil {
		return
	}
	s = string(buf)
	if r.cache == nil {
		r.cache = make(map[uint32]string)
	}
	r.cache[id] = s
	return
}

func (r *dbReader) readByte(pos int64) (byte, error) {
	buf := [1]byte{}
	_, err := r.ReadAt(buf[:], pos)
	return buf[0], err
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
}

const zeroPrefix = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xFF\xFF"

func NormalizeIP(ip net.IP) net.IP {
	if len(ip) == net.IPv4len {
		return ip
	}
	if len(ip) == net.IPv6len && string(ip[:12]) == zeroPrefix {
		return ip[12:16]
	}
	return ip
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
