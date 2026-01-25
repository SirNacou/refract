package geo

import (
	"log/slog"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

type GeoInfo struct {
	CountryCode *string
	CountryName *string
	City        *string
	Latitude    *float64
	Longitude   *float64
}

type GeoLookup struct {
	reader *maxminddb.Reader
}

func NewGeoLookup(dbPath string) (*GeoLookup, error) {
	reader, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &GeoLookup{reader: reader}, nil
}

func (g *GeoLookup) Close() error {
	return g.reader.Close()
}

// Lookup performs GeoIP lookup for the given IP address
// Returns nil for private/invalid IPs without logging (expected)
// Logs warnings only for lookup failures on public IPs
func (g *GeoLookup) Lookup(ip string) *GeoInfo {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil
	}

	// Skip lookup for private IPs (matches redirector behavior)
	if isPrivateIP(addr) {
		return nil
	}

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
			Names   map[string]string
		} `maxminddb:"country"`
		City struct {
			Names map[string]string
		} `maxminddb:"city"`
		Location struct {
			Latitude  float64 `maxminddb:"latitude"`
			Longitude float64 `maxminddb:"longitude"`
		} `maxminddb:"location"`
	}

	err = g.reader.Lookup(addr).Decode(&record)
	if err != nil {
		slog.Warn("GeoIP lookup failed", "ip", ip, "error", err)
		return nil
	}

	geoInfo := &GeoInfo{}

	if record.Country.ISOCode != "" {
		geoInfo.CountryCode = &record.Country.ISOCode
	}
	if name, ok := record.Country.Names["en"]; ok {
		geoInfo.CountryName = &name
	}
	if name, ok := record.City.Names["en"]; ok {
		geoInfo.City = &name
	}

	// Always set coordinates if lookup succeeded (even if zero)
	// MaxMind returns 0,0 for missing data, but we can't distinguish from actual 0,0
	// Setting them allows downstream to decide whether to use
	geoInfo.Latitude = &record.Location.Latitude
	geoInfo.Longitude = &record.Location.Longitude

	return geoInfo
}

// isPrivateIP checks if an IP address is private/internal (matches redirector logic)
// Private IPs are not looked up to avoid unnecessary warnings
func isPrivateIP(addr netip.Addr) bool {
	if addr.Is4() {
		// IPv4: private, loopback, link-local, unspecified
		return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsUnspecified()
	}
	// IPv6: loopback, unspecified (matches redirector - does not include ULA fc00::/7)
	return addr.IsLoopback() || addr.IsUnspecified()
}
