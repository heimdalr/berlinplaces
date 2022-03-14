package places

import "encoding/json"

type District struct {
	Postcode string
	District string
}

type DistrictMap map[string]*District

// Class enumerates different place types / classes.
type Class int32

const (

	// StreetClass is the place class of streets.
	StreetClass = iota

	// LocationClass is place class of locations.
	LocationClass

	// HouseNumberClass is the place class of house numbers / buildings.
	HouseNumberClass
)

// String implements the stringer interface for Class.
func (c Class) String() string {
	return [...]string{"street", "location", "houseNumber"}[c]
}

type Place struct {
	ID           int64
	Class        Class
	Type         string
	Name         string
	Street       *Place
	HouseNumber  string
	District     *District
	Length       int
	Lat          float64
	Lon          float64
	Relevance    uint64
	SimpleName   string
	HouseNumbers []*Place
}

type PlaceMap map[int64]*Place

// MarshalJSON marshall a place to JSON.
func (p *Place) MarshalJSON() ([]byte, error) {
	var (
		streetName, postcode, district string
		streetID                       *int64
		length                         *int
	)
	if p.District != nil {
		postcode = p.District.Postcode
		district = p.District.District
	}
	switch p.Class {
	case StreetClass:
		length = &p.Length
	case LocationClass:
		streetName = p.Street.Name
		streetID = &p.Street.ID
	default: // HouseNumberClass
		streetName = p.Street.Name
		streetID = &p.Street.ID
	}
	return json.Marshal(&struct {
		ID          int64   `json:"id"`
		Class       string  `json:"class"`
		Type        string  `json:"type,omitempty"`
		Name        string  `json:"name,omitempty"`
		Street      string  `json:"street,omitempty"`
		StreetID    *int64  `json:"streetID,omitempty"`
		HouseNumber string  `json:"houseNumber,omitempty"`
		Postcode    string  `json:"postcode"`
		District    string  `json:"district"`
		Length      *int    `json:"length,omitempty"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Relevance   uint64  `json:"relevance"`
	}{
		ID:          p.ID,
		Class:       p.Class.String(),
		Type:        p.Type,
		Name:        p.Name,
		Street:      streetName,
		StreetID:    streetID,
		HouseNumber: p.HouseNumber,
		Postcode:    postcode,
		District:    district,
		Length:      length,
		Lat:         p.Lat,
		Lon:         p.Lon,
		Relevance:   p.Relevance,
	})
}

// placeLesser compares two places wrt. string length and lexical order.
// placeLesser return true, if the first place is less than the second one.
func placeLesser(i, j *Place) bool {

	// less by character length
	if len(i.SimpleName) != len(j.SimpleName) {
		if len(i.SimpleName) < len(j.SimpleName) {
			return true
		} else {
			return false
		}
	}

	// less by lex
	if i.SimpleName != j.SimpleName {
		if i.SimpleName < j.SimpleName {
			return true
		} else {
			return false
		}
	}

	// defaults
	return false
}

// deDuplicate removes duplicate places.
func deDuplicate(results []*Place) []*Place {

	ids := make(map[int64]interface{})
	var places []*Place

	for _, p := range results {
		if _, exists := ids[p.ID]; !exists {
			ids[p.ID] = struct{}{}
			places = append(places, p)
		}
	}

	return places
}
