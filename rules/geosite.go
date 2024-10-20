package rules

type GeoSiteRule struct {
	country    string
	adapter    string
	recodeSize int
}

func (g *GeoSiteRule) Country() string {
	return g.country
}
