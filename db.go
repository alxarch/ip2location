package ip2location

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"net"
	"time"
)

type DB struct {
	kind EntryKind
	date time.Time
	v4   *dbEntryIndex
	v6   *dbEntryIndex
}

func Open(path string, fields ...Field) (*DB, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	kind, date, err := readDBMeta(r)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		fields, err = requireFields(kind, fields)
		if err != nil {
			return nil, err
		}
	} else {
		fields = dbFields[kind]
	}

	db := dbReader{
		ReaderAt: r,
		kind:     kind,
		fields:   fields,
		cache:    make(map[uint32]string),
	}
	v4, err := db.readEntryIndex(4)
	if err != nil {
		return nil, err
	}
	v6, err := db.readEntryIndex(6)
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
	ips     []byte
	entries []Entry
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

func (db *dbEntries) entry(i int) *Entry {
	if 0 <= i && i < len(db.entries) {
		return &db.entries[i]
	}
	return nil
}

func (db *dbEntries) Lookup(ip net.IP) *Entry {
	lo, hi := 0, len(db.entries)
	for lo <= hi {
		mid := (lo + hi) / 2
		if compareAt(ip, db.ips, mid) < 0 {
			hi = mid - 1
		} else if lo = mid + 1; compareAt(ip, db.ips, lo) < 0 {
			return db.entry(mid)
		}
	}
	return nil
}

func (db *dbEntries) append(ip net.IP, e *Entry) {
	db.ips = append(db.ips, ip...)
	db.entries = append(db.entries, *e)
}

type dbEntryIndex struct {
	index [256]dbEntries
}

func (db *dbEntryIndex) Lookup(ip net.IP) *Entry {
	if db != nil && len(ip) > 0 {
		idx, ip := &db.index[ip[0]], ip[1:]
		return idx.Lookup(ip)
	}
	return nil
}

func (r dbReader) readEntryIndex(v int) (*dbEntryIndex, error) {
	iter, err := newRowIterator(r, v, r.kind)
	if err != nil || iter == nil {
		return nil, err
	}
	db := new(dbEntryIndex)
	for iter.Next() {
		row := iter.Row()
		ip, data := rowIP(row, v)
		if len(ip) > 0 {
			k, ip := ip[0], ip[1:]
			e := Entry{}
			if err := r.readEntry(&e, data); err != nil {
				return nil, err
			}
			db.index[k].append(ip, &e)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return db, nil
}

type dbReader struct {
	io.ReaderAt
	kind   EntryKind
	fields Fields
	cache  map[uint32]string
}

func (r *dbReader) readField(e *Entry, f Field, n uint32) (err error) {
	switch f {
	case FieldCountry:
		e.Country, err = r.readString(n)
	case FieldRegion:
		e.Region, err = r.readString(n)
	case FieldCity:
		e.City, err = r.readString(n)
	case FieldISP:
		e.ISP, err = r.readString(n)
	case FieldLatitude:
		e.Latitude = math.Float32frombits(n)
	case FieldLongitude:
		e.Longitude = math.Float32frombits(n)
	case FieldDomain:
		e.Domain, err = r.readString(n)
	case FieldZipCode:
		e.ZipCode, err = r.readString(n)
	case FieldTimeZone:
		e.TimeZone, err = r.readString(n)
	case FieldNetSpeed:
		e.NetSpeed, err = r.readString(n)
	case FieldIDDCode:
		e.IDDCode, err = r.readString(n)
	case FieldAreaCode:
		e.AreaCode, err = r.readString(n)
	case FieldWeatherCode:
		e.WeatherStationCode, err = r.readString(n)
	case FieldWeatherName:
		e.WeatherStationName, err = r.readString(n)
	case FieldMCC:
		e.MCC, err = r.readString(n)
	case FieldMNC:
		e.MCC, err = r.readString(n)
	case FieldMobileBrand:
		e.MobileBrand, err = r.readString(n)
	case FieldElevation:
		e.Elevation = math.Float32frombits(n)
	case FieldUsageType:
		e.UsageType, err = r.readString(n)
	}
	return

}
func (r *dbReader) readEntry(e *Entry, buf []byte) error {
	var n uint32
	for _, f := range r.fields {
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
	size, err := readByte(r.ReaderAt, pos)
	if err != nil {
		return
	}
	buf := make([]byte, size)
	_, err = r.ReadAt(buf[:], pos+1)
	if err != nil {
		return
	}
	s = string(buf)
	// if r.cache == nil {
	// 	r.cache = make(map[uint32]string)
	// }
	r.cache[id] = s
	return
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
