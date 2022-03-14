package data

type District struct {
	Postcode string
	District string
}

type Street struct {
	ID       string
	Name     string
	Postcode string
	Lat      float64
	Lon      float64
	Length   int
}

type Location struct {
	ID          string
	Type        string
	Name        string
	StreetID    string
	HouseNumber string
	Postcode    string
	Lat         float64
	Lon         float64
}

type HouseNumber struct {
	ID          string
	StreetID    string
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
