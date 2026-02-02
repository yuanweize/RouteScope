package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

// Supported language codes
const (
	LangEnglish = "en"
	LangChinese = "zh-CN"
)

type Location struct {
	City      string  `json:"city"`
	CityEN    string  `json:"city_en"`   // English name for localization
	Subdiv    string  `json:"subdiv"`    // Province/State
	SubdivEN  string  `json:"subdiv_en"` // English subdivision
	Country   string  `json:"country"`
	CountryEN string  `json:"country_en"` // English country
	ISOCode   string  `json:"iso_code"`
	ISP       string  `json:"isp"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Precision string  `json:"precision"` // "city", "subdivision", "country", "none"
}

// MaxMindCityRecord represents MaxMind GeoLite2-City structure
type MaxMindCityRecord struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	Country struct {
		Names   map[string]string `maxminddb:"names"`
		IsoCode string            `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

// DBIPCityRecord represents DB-IP City Lite structure
// DB-IP uses flat string fields instead of nested names maps
type DBIPCityRecord struct {
	City        string  `maxminddb:"city"`
	State1      string  `maxminddb:"state1"`      // Province/State
	State2      string  `maxminddb:"state2"`      // Sub-region (optional)
	CountryCode string  `maxminddb:"country_code"`
	Latitude    float64 `maxminddb:"latitude"`
	Longitude   float64 `maxminddb:"longitude"`
	Postcode    string  `maxminddb:"postcode"`
	Timezone    string  `maxminddb:"timezone"`
}

// Country code to Chinese name mapping for common countries
var countryCodeToChinese = map[string]string{
	"CN": "中国", "US": "美国", "JP": "日本", "KR": "韩国",
	"DE": "德国", "FR": "法国", "GB": "英国", "RU": "俄罗斯",
	"SG": "新加坡", "HK": "香港", "TW": "台湾", "AU": "澳大利亚",
	"CA": "加拿大", "NL": "荷兰", "IN": "印度", "BR": "巴西",
}

type Provider struct {
	cityDB   *maxminddb.Reader // Use maxminddb for better compatibility
	ispDB    *geoip2.Reader    // Keep geoip2 for ISP (standard MaxMind format)
	dbType   string            // "maxmind" or "dbip"
}

func NewProvider(cityDBPath, ispDBPath string) (*Provider, error) {
	p := &Provider{}

	if cityDBPath != "" {
		db, err := maxminddb.Open(cityDBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open City DB: %w", err)
		}
		p.cityDB = db
		
		// Detect database type by metadata
		meta := db.Metadata
		if meta.DatabaseType == "GeoLite2-City" || meta.DatabaseType == "GeoIP2-City" {
			p.dbType = "maxmind"
		} else {
			// DB-IP and other databases
			p.dbType = "dbip"
		}
	}

	if ispDBPath != "" {
		db, err := geoip2.Open(ispDBPath)
		if err != nil {
			// ISP DB is optional, just log or ignore
		} else {
			p.ispDB = db
		}
	}

	return p, nil
}

// Lookup returns location data with both Chinese and English names
// The primary fields (City, Subdiv, Country) use Chinese if available, else English
// The *EN fields always contain English names for API consumers to choose
func (p *Provider) Lookup(ipStr string) (*Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", ipStr)
	}

	loc := &Location{}

	if p.cityDB != nil {
		if p.dbType == "dbip" {
			// DB-IP City Lite format
			var record DBIPCityRecord
			err := p.cityDB.Lookup(ip, &record)
			if err == nil {
				// DB-IP uses English names directly
				loc.CityEN = record.City
				loc.City = record.City // No Chinese in DB-IP
				
				loc.SubdivEN = record.State1
				loc.Subdiv = record.State1
				
				loc.ISOCode = record.CountryCode
				loc.CountryEN = record.CountryCode
				// Try to get Chinese country name
				if cn, ok := countryCodeToChinese[record.CountryCode]; ok {
					loc.Country = cn
				} else {
					loc.Country = record.CountryCode
				}
				
				loc.Latitude = record.Latitude
				loc.Longitude = record.Longitude
			}
		} else {
			// MaxMind GeoLite2-City format
			var record MaxMindCityRecord
			err := p.cityDB.Lookup(ip, &record)
			if err == nil {
				// Get both language versions for each field
				// City
				loc.CityEN = record.City.Names["en"]
				loc.City = record.City.Names["zh-CN"]
				if loc.City == "" {
					loc.City = loc.CityEN // Fallback to English
				}

				// Subdivision (Province/State)
				if len(record.Subdivisions) > 0 {
					loc.SubdivEN = record.Subdivisions[0].Names["en"]
					loc.Subdiv = record.Subdivisions[0].Names["zh-CN"]
					if loc.Subdiv == "" {
						loc.Subdiv = loc.SubdivEN
					}
				}

				// Country
				loc.CountryEN = record.Country.Names["en"]
				loc.Country = record.Country.Names["zh-CN"]
				if loc.Country == "" {
					loc.Country = loc.CountryEN
				}

				loc.ISOCode = record.Country.IsoCode
				loc.Latitude = record.Location.Latitude
				loc.Longitude = record.Location.Longitude
			}
		}

		// Determine precision level
		if loc.City != "" || loc.CityEN != "" {
			loc.Precision = "city"
		} else if loc.Subdiv != "" || loc.SubdivEN != "" {
			loc.Precision = "subdivision"
		} else if loc.Country != "" || loc.CountryEN != "" {
			loc.Precision = "country"
		} else {
			loc.Precision = "none"
		}
	}

	if p.ispDB != nil {
		record, err := p.ispDB.ISP(ip)
		if err == nil {
			loc.ISP = record.Organization
			if loc.ISP == "" {
				loc.ISP = record.ISP
			}
		}
	}

	return loc, nil
}

func (p *Provider) Close() {
	if p.cityDB != nil {
		p.cityDB.Close()
	}
	if p.ispDB != nil {
		p.ispDB.Close()
	}
}
