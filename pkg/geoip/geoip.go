package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type Location struct {
	City       string
	Subdiv     string // Province/State
	Country    string
	ISOCode    string
	ISP        string
	Latitude   float64
	Longitude  float64
	Precision  string // "city", "subdivision", "country", "none"
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

func (p *Provider) Lookup(ipStr string) (*Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", ipStr)
	}

	loc := &Location{}

	if p.cityDB != nil {
		record, err := p.cityDB.City(ip)
		if err == nil {
			// Try City first (highest precision)
			loc.City = record.City.Names["zh-CN"]
			if loc.City == "" {
				loc.City = record.City.Names["en"]
			}

			// Try Subdivision (Province/State) as fallback
			if len(record.Subdivisions) > 0 {
				loc.Subdiv = record.Subdivisions[0].Names["zh-CN"]
				if loc.Subdiv == "" {
					loc.Subdiv = record.Subdivisions[0].Names["en"]
				}
			}

			loc.Country = record.Country.Names["zh-CN"]
			if loc.Country == "" {
				loc.Country = record.Country.Names["en"]
			}
			loc.ISOCode = record.Country.IsoCode
			loc.Latitude = record.Location.Latitude
			loc.Longitude = record.Location.Longitude

			// Determine precision level
			if loc.City != "" {
				loc.Precision = "city"
			} else if loc.Subdiv != "" {
				loc.Precision = "subdivision"
			} else if loc.Country != "" {
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
