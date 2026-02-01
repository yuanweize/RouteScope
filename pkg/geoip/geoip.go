package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
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

type Provider struct {
	cityDB *geoip2.Reader
	ispDB  *geoip2.Reader
}

func NewProvider(cityDBPath, ispDBPath string) (*Provider, error) {
	p := &Provider{}

	if cityDBPath != "" {
		db, err := geoip2.Open(cityDBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open City DB: %w", err)
		}
		p.cityDB = db
	}

	if ispDBPath != "" {
		db, err := geoip2.Open(ispDBPath)
		if err != nil {
			// ISP DB is optional, just log or ignore?
			// For this tool, it's fine to be optional
		}
		p.ispDB = db
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
		record, err := p.cityDB.City(ip)
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
