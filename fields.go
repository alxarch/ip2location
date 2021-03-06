package ip2location

type EntryKind uint8

func (k EntryKind) Fields() (fields Fields) {
	return append(fields, dbFields[k]...)
}

type Entry struct {
	Country            string
	Region             string
	City               string
	ISP                string
	Latitude           float32
	Longitude          float32
	Domain             string
	ZipCode            string
	TimeZone           string
	NetSpeed           string
	IDDCode            string
	AreaCode           string
	WeatherStationCode string
	WeatherStationName string
	MCC                string
	MNC                string
	MobileBrand        string
	Elevation          float32
	UsageType          string
}

type Field uint

const (
	FieldCountry Field = 1 << iota
	FieldRegion
	FieldCity
	FieldISP
	FieldLatitude
	FieldLongitude
	FieldDomain
	FieldZipCode
	FieldTimeZone
	FieldNetSpeed
	FieldIDDCode
	FieldAreaCode
	FieldWeatherCode
	FieldWeatherName
	FieldMCC
	FieldMNC
	FieldMobileBrand
	FieldElevation
	FieldUsageType
	maxField
)

var fieldNames = map[Field]string{
	FieldCountry:     "Country",
	FieldRegion:      "Region",
	FieldCity:        "City",
	FieldISP:         "ISP",
	FieldLatitude:    "Latitude",
	FieldLongitude:   "Longitude",
	FieldDomain:      "Domain",
	FieldZipCode:     "ZipCode",
	FieldTimeZone:    "TimeZone",
	FieldNetSpeed:    "NetSpeed",
	FieldIDDCode:     "IDDCode",
	FieldAreaCode:    "AreaCode",
	FieldWeatherCode: "WeatherCode",
	FieldWeatherName: "WeatherName",
	FieldMCC:         "MCC",
	FieldMNC:         "MNC",
	FieldMobileBrand: "MobileBrand",
	FieldElevation:   "Elevation",
	FieldUsageType:   "UsageType",
}

func (f Field) String() string {
	return fieldNames[f]
}

var dbFields = map[EntryKind][]Field{
	1: []Field{
		FieldCountry,
	},
	2: []Field{
		FieldCountry,
		FieldISP,
	},
	3: []Field{
		FieldCountry, FieldRegion, FieldCity,
	},
	4: []Field{
		FieldCountry, FieldRegion, FieldCity,
		FieldISP,
	},
	5: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude,
	},
	6: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude,
		FieldISP,
	},
	7: []Field{FieldCountry, FieldRegion, FieldCity,
		FieldISP, FieldDomain,
	},
	8: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude,
		FieldISP, FieldDomain,
	},
	9: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode,
	},
	10: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode,
		FieldISP, FieldDomain,
	},
	11: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
	},
	12: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
	},
	13: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldTimeZone,
		FieldNetSpeed,
	},
	14: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
	},
	15: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldIDDCode, FieldAreaCode,
	},
	16: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
		FieldIDDCode, FieldAreaCode,
	},
	17: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldTimeZone,
		FieldNetSpeed,
		FieldWeatherCode, FieldWeatherName,
	},
	18: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
		FieldIDDCode, FieldAreaCode,
		FieldWeatherCode, FieldWeatherName,
	},
	19: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude,
		FieldISP, FieldDomain,
		FieldIDDCode, FieldMCC, FieldMNC, FieldMobileBrand,
	},
	20: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
		FieldIDDCode, FieldAreaCode,
		FieldWeatherCode, FieldWeatherName,
		FieldMCC, FieldMNC, FieldMobileBrand,
	},
	21: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldAreaCode,
		FieldElevation,
	},
	22: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
		FieldIDDCode, FieldAreaCode,
		FieldWeatherCode, FieldWeatherName,
		FieldMCC, FieldMNC, FieldMobileBrand,
		FieldElevation,
	},
	23: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude,
		FieldISP, FieldDomain,
		FieldMCC, FieldMNC, FieldMobileBrand,
		FieldUsageType,
	},
	24: []Field{
		FieldCountry, FieldRegion, FieldCity, FieldLatitude, FieldLongitude, FieldZipCode, FieldTimeZone,
		FieldISP, FieldDomain,
		FieldNetSpeed,
		FieldIDDCode, FieldAreaCode,
		FieldWeatherCode, FieldWeatherName,
		FieldMCC, FieldMNC, FieldMobileBrand,
		FieldElevation,
		FieldUsageType,
	},
}

type Fields []Field

func (fields Fields) IndexOf(f Field) int {
	for i, ff := range fields {
		if ff == f {
			return i
		}
	}
	return -1
}
func (fields Fields) Copy() (cp Fields) {
	return append(cp, fields...)
}

type FieldMask uint

func (fields Fields) Mask() (m FieldMask) {
	for _, f := range fields {
		m |= FieldMask(f)
	}
	return
}

const (
	allFields FieldMask = FieldMask(FieldCountry | FieldRegion | FieldCity | FieldLatitude | FieldLongitude | FieldZipCode | FieldTimeZone | FieldISP | FieldDomain | FieldNetSpeed | FieldIDDCode | FieldAreaCode | FieldWeatherCode | FieldWeatherName | FieldMCC | FieldMNC | FieldMobileBrand | FieldElevation | FieldUsageType)
)

func (m FieldMask) Has(f Field) bool {
	return Field(m)&f == f
}

const (
	UsageCDN          = "CDN"
	UsageCommercial   = "COM"
	UsageDataCenter   = "DCH"
	UsageEducation    = "EDU"
	UsageGoverment    = "GOV"
	UsageISP          = "ISP"
	UsageMobile       = "MOB"
	UsageLibrary      = "LIB"
	UsageMilitary     = "MIL"
	UsageOrganization = "ORG"
	UsageReserved     = "RSV"
	UsageSearchEngine = "SES"
	UsageISPMobile    = "ISP/MOB"
)
