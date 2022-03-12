package data

type District struct {
	Postcode string `csv:"postcode"`
	District string `csv:"district"`
}

type Street struct {
	ID       int     `csv:"id"`
	Name     string  `csv:"name"`
	Cluster  string  `csv:"cluster"`
	Postcode string  `csv:"postcode"`
	Lat      float64 `csv:"lat"`
	Lon      float64 `csv:"lon"`
	Length   int     `csv:"length"`
}

type Location struct {
	Type        string  `csv:"type"`
	Name        string  `csv:"name"`
	StreetID    int     `csv:"street_id"`
	HouseNumber string  `csv:"house_number"`
	Postcode    string  `csv:"postcode"`
	Lat         float64 `csv:"lat"`
	Lon         float64 `csv:"lon"`
}

type HouseNumber struct {
	StreetID    int     `csv:"street_id"`
	HouseNumber string  `csv:"house_number"`
	Postcode    string  `csv:"postcode"`
	Lat         float64 `csv:"lat"`
	Lon         float64 `csv:"lon"`
}

type Data struct {
	Districts    []*District
	Streets      []*Street
	Locations    []*Location
	HouseNumbers []*HouseNumber
}

type Provider interface {
	Get() (*Data, error)
}
