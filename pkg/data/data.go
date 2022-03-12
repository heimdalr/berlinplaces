package data

type District struct {
	Postcode string
	District string
}

type Street struct {
	ID       int
	Name     string
	Cluster  string
	Postcode string
	Lat      float64
	Lon      float64
	Length   int
}

type Location struct {
	Type        string
	Name        string
	StreetID    int
	HouseNumber string
	Postcode    string
	Lat         float64
	Lon         float64
}

type HouseNumber struct {
	StreetID    int
	HouseNumber string
	Postcode    string
	Lat         float64
	Lon         float64
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
