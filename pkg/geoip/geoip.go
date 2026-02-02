package geoip

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

// Embed China city coordinates data (from AreaCity-JsSpider-StatsGov)
// Source: https://github.com/xiangyuecn/AreaCity-JsSpider-StatsGov
//
//go:embed china_cities_geo.json
var chinaCitiesGeoJSON []byte

// chinaCityCoords stores city name -> [lat, lng]
var chinaCityCoords map[string][2]float64
var chinaCityCoordsOnce sync.Once

// loadChinaCityCoords loads the embedded China city coordinates
func loadChinaCityCoords() {
	chinaCityCoordsOnce.Do(func() {
		chinaCityCoords = make(map[string][2]float64)
		if err := json.Unmarshal(chinaCitiesGeoJSON, &chinaCityCoords); err != nil {
			log.Printf("[GeoIP] Warning: failed to load China city coordinates: %v", err)
			return
		}
		log.Printf("[GeoIP] Loaded %d China city coordinates", len(chinaCityCoords))
	})
}

// lookupChinaCityCoords looks up coordinates for a Chinese city
// Returns lat, lng, found
func lookupChinaCityCoords(city, subdiv string) (float64, float64, bool) {
	loadChinaCityCoords()
	if chinaCityCoords == nil {
		return 0, 0, false
	}

	// Try exact city name first (e.g., "湘潭")
	if city != "" {
		// Try with "市" suffix
		if coords, ok := chinaCityCoords[city+"市"]; ok {
			return coords[0], coords[1], true
		}
		// Try exact match
		if coords, ok := chinaCityCoords[city]; ok {
			return coords[0], coords[1], true
		}
	}

	// Fallback to subdivision (province) center
	if subdiv != "" {
		// Try with "省" suffix
		if coords, ok := chinaCityCoords[subdiv+"省"]; ok {
			return coords[0], coords[1], true
		}
		if coords, ok := chinaCityCoords[subdiv]; ok {
			return coords[0], coords[1], true
		}
	}

	return 0, 0, false
}

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
type DBIPCityRecord struct {
	City        string  `maxminddb:"city"`
	State1      string  `maxminddb:"state1"`
	State2      string  `maxminddb:"state2"`
	CountryCode string  `maxminddb:"country_code"`
	Latitude    float64 `maxminddb:"latitude"`
	Longitude   float64 `maxminddb:"longitude"`
	Postcode    string  `maxminddb:"postcode"`
	Timezone    string  `maxminddb:"timezone"`
}

// Country code to Chinese name mapping
var countryCodeToChinese = map[string]string{
	"CN": "中国", "US": "美国", "JP": "日本", "KR": "韩国",
	"DE": "德国", "FR": "法国", "GB": "英国", "RU": "俄罗斯",
	"SG": "新加坡", "HK": "香港", "TW": "台湾", "AU": "澳大利亚",
	"CA": "加拿大", "NL": "荷兰", "IN": "印度", "BR": "巴西",
	"CZ": "捷克", "PL": "波兰", "IT": "意大利", "ES": "西班牙",
}

// China province Chinese to English mapping
var chinaProvinceToEN = map[string]string{
	"北京": "Beijing", "天津": "Tianjin", "河北": "Hebei", "山西": "Shanxi",
	"内蒙古": "Inner Mongolia", "辽宁": "Liaoning", "吉林": "Jilin", "黑龙江": "Heilongjiang",
	"上海": "Shanghai", "江苏": "Jiangsu", "浙江": "Zhejiang", "安徽": "Anhui",
	"福建": "Fujian", "江西": "Jiangxi", "山东": "Shandong", "河南": "Henan",
	"湖北": "Hubei", "湖南": "Hunan", "广东": "Guangdong", "广西": "Guangxi",
	"海南": "Hainan", "重庆": "Chongqing", "四川": "Sichuan", "贵州": "Guizhou",
	"云南": "Yunnan", "西藏": "Tibet", "陕西": "Shaanxi", "甘肃": "Gansu",
	"青海": "Qinghai", "宁夏": "Ningxia", "新疆": "Xinjiang", "香港": "Hong Kong",
	"澳门": "Macau", "台湾": "Taiwan",
}

// Convert Chinese province/city name to English (pinyin-based)
func toEnglishName(zhName string) string {
	// Try province mapping first
	if en, ok := chinaProvinceToEN[zhName]; ok {
		return en
	}
	// Remove common suffixes and try again
	name := strings.TrimSuffix(zhName, "省")
	name = strings.TrimSuffix(name, "市")
	name = strings.TrimSuffix(name, "自治区")
	name = strings.TrimSuffix(name, "特别行政区")
	if en, ok := chinaProvinceToEN[name]; ok {
		return en
	}
	// For cities, return as-is (pinyin would require external library)
	return zhName
}

type Provider struct {
	cityDB      *maxminddb.Reader // DB-IP/MaxMind for non-China IPs
	ispDB       *geoip2.Reader    // MaxMind ISP database
	ip2regionDB *xdb.Searcher     // ip2region for China IPs (high precision)
	dbType      string            // "maxmind" or "dbip"
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
			p.dbType = "dbip"
		}
	}

	if ispDBPath != "" {
		db, err := geoip2.Open(ispDBPath)
		if err == nil {
			p.ispDB = db
		}
	}

	return p, nil
}

// LoadIP2Region loads the ip2region database for high-precision China IP lookup
func (p *Provider) LoadIP2Region(xdbPath string) error {
	// Load entire xdb file into memory for best performance
	cBuff, err := xdb.LoadContentFromFile(xdbPath)
	if err != nil {
		return fmt.Errorf("failed to load ip2region xdb: %w", err)
	}

	// Get version from header
	header, err := xdb.LoadHeaderFromBuff(cBuff)
	if err != nil {
		return fmt.Errorf("failed to load ip2region header: %w", err)
	}

	version, err := xdb.VersionFromHeader(header)
	if err != nil {
		return fmt.Errorf("failed to get ip2region version: %w", err)
	}

	searcher, err := xdb.NewWithBuffer(version, cBuff)
	if err != nil {
		return fmt.Errorf("failed to create ip2region searcher: %w", err)
	}

	p.ip2regionDB = searcher
	log.Printf("[GeoIP] ip2region database loaded for high-precision China IP lookup")
	return nil
}

// Lookup returns location data with both Chinese and English names
// For China IPs, uses ip2region for high precision (city + ISP)
// For other IPs, uses DB-IP/MaxMind database
func (p *Provider) Lookup(ipStr string) (*Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", ipStr)
	}

	loc := &Location{}

	// Try ip2region first for China IPs (higher precision)
	if p.ip2regionDB != nil {
		region, err := p.ip2regionDB.SearchByStr(ipStr)
		if err == nil && region != "" {
			// ip2region format: 国家|省份|城市|ISP|iso-code
			// Example: 中国|湖南|湘潭|电信|CN
			parts := strings.Split(region, "|")
			if len(parts) >= 4 && parts[0] == "中国" {
				// This is a China IP with detailed info
				loc.Country = parts[0]
				loc.CountryEN = "China"
				loc.ISOCode = "CN"

				if parts[1] != "0" && parts[1] != "" {
					loc.Subdiv = parts[1]
					loc.SubdivEN = toEnglishName(parts[1])
				}

				if parts[2] != "0" && parts[2] != "" {
					loc.City = parts[2]
					loc.CityEN = toEnglishName(parts[2])
				}

				if parts[3] != "0" && parts[3] != "" {
					loc.ISP = parts[3]
				}

				// Set precision
				if loc.City != "" {
					loc.Precision = "city"
				} else if loc.Subdiv != "" {
					loc.Precision = "subdivision"
				} else {
					loc.Precision = "country"
				}

				// Lookup coordinates for China city
				if lat, lng, ok := lookupChinaCityCoords(loc.City, loc.Subdiv); ok {
					loc.Latitude = lat
					loc.Longitude = lng
				}

				return loc, nil
			}
		}
	}

	// Fallback to DB-IP/MaxMind for non-China IPs
	if p.cityDB != nil {
		if p.dbType == "dbip" {
			var record DBIPCityRecord
			err := p.cityDB.Lookup(ip, &record)
			if err == nil {
				loc.CityEN = record.City
				loc.City = record.City

				loc.SubdivEN = record.State1
				loc.Subdiv = record.State1

				loc.ISOCode = record.CountryCode
				loc.CountryEN = record.CountryCode
				if cn, ok := countryCodeToChinese[record.CountryCode]; ok {
					loc.Country = cn
				} else {
					loc.Country = record.CountryCode
				}

				loc.Latitude = record.Latitude
				loc.Longitude = record.Longitude
			}
		} else {
			var record MaxMindCityRecord
			err := p.cityDB.Lookup(ip, &record)
			if err == nil {
				loc.CityEN = record.City.Names["en"]
				loc.City = record.City.Names["zh-CN"]
				if loc.City == "" {
					loc.City = loc.CityEN
				}

				if len(record.Subdivisions) > 0 {
					loc.SubdivEN = record.Subdivisions[0].Names["en"]
					loc.Subdiv = record.Subdivisions[0].Names["zh-CN"]
					if loc.Subdiv == "" {
						loc.Subdiv = loc.SubdivEN
					}
				}

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

	// ISP lookup from MaxMind ISP database (for non-China IPs)
	if p.ispDB != nil && loc.ISP == "" {
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
	if p.ip2regionDB != nil {
		p.ip2regionDB.Close()
	}
}
