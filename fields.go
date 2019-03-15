package ip2location

type EntryKind uint8

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

type field uint

const (
	fieldCountry field = iota
	fieldRegion
	fieldCity
	fieldISP
	fieldLatitude
	fieldLongitude
	fieldDomain
	fieldZipCode
	fieldTimeZone
	fieldNetSpeed
	fieldIDDCode
	fieldAreaCode
	fieldWeatherCode
	fieldWeatherName
	fieldMCC
	fieldMNC
	fieldMobileBrand
	fieldElevation
	fieldUsageType
	maxField
)

var dbFields = map[EntryKind][]field{
	1: []field{
		fieldCountry,
	},
	2: []field{
		fieldCountry,
		fieldISP,
	},
	3: []field{
		fieldCountry, fieldRegion, fieldCity,
	},
	4: []field{
		fieldCountry, fieldRegion, fieldCity,
		fieldISP,
	},
	5: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude,
	},
	6: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude,
		fieldISP,
	},
	7: []field{fieldCountry, fieldRegion, fieldCity,
		fieldISP, fieldDomain,
	},
	8: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude,
		fieldISP, fieldDomain,
	},
	9: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode,
	},
	10: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode,
		fieldISP, fieldDomain,
	},
	11: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
	},
	12: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
	},
	13: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldTimeZone,
		fieldNetSpeed,
	},
	14: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
	},
	15: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldIDDCode, fieldAreaCode,
	},
	16: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
		fieldIDDCode, fieldAreaCode,
	},
	17: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldTimeZone,
		fieldNetSpeed,
		fieldWeatherCode, fieldWeatherName,
	},
	18: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
		fieldIDDCode, fieldAreaCode,
		fieldWeatherCode, fieldWeatherName,
	},
	19: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude,
		fieldISP, fieldDomain,
		fieldIDDCode, fieldMCC, fieldMNC, fieldMobileBrand,
	},
	20: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
		fieldIDDCode, fieldAreaCode,
		fieldWeatherCode, fieldWeatherName,
		fieldMCC, fieldMNC, fieldMobileBrand,
	},
	21: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldAreaCode,
		fieldElevation,
	},
	22: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
		fieldIDDCode, fieldAreaCode,
		fieldWeatherCode, fieldWeatherName,
		fieldMCC, fieldMNC, fieldMobileBrand,
		fieldElevation,
	},
	23: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude,
		fieldISP, fieldDomain,
		fieldMCC, fieldMNC, fieldMobileBrand,
		fieldUsageType,
	},
	24: []field{
		fieldCountry, fieldRegion, fieldCity, fieldLatitude, fieldLongitude, fieldZipCode, fieldTimeZone,
		fieldISP, fieldDomain,
		fieldNetSpeed,
		fieldIDDCode, fieldAreaCode,
		fieldWeatherCode, fieldWeatherName,
		fieldMCC, fieldMNC, fieldMobileBrand,
		fieldElevation,
		fieldUsageType,
	},
}
