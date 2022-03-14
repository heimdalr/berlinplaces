package data

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"io"
	"strings"
)

type CSVPlace struct {
	ID          int64
	Type        string
	Name        string
	StreetID    int64
	HouseNumber string
	Postcode    string
	Length      int
	Lat         float64
	Lon         float64
}

type CSVProvider struct {
	DistrictsReader io.Reader
	PlacesReader    io.Reader
}

// Get implements the Provider interface for CSVProvider.
func (provider CSVProvider) Get() (places.DistrictMap, places.PlaceMap, *places.Metrics, error) {

	// normalizer is the function to apply to headers and struct fields before trying to match.
	// see: https://pkg.go.dev/github.com/gocarina/gocsv#Normalizer
	gocsv.SetHeaderNormalizer(func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", "")
	})

	districtsMap := make(places.DistrictMap)

	// unmarshall districts into map
	districtsChan := make(chan *places.District)
	districtsDoneChan := make(chan bool)
	go func() {
		for district := range districtsChan {
			districtsMap[district.Postcode] = district
		}
		districtsDoneChan <- true
	}()
	if err := gocsv.UnmarshalToChan(provider.DistrictsReader, districtsChan); err != nil {
		return nil, nil, nil, err
	}
	<-districtsDoneChan

	// unmarshall places into map
	placeMap := make(places.PlaceMap)
	counts := make(map[places.Class]int32)
	placesChan := make(chan CSVPlace)
	placesDoneChan := make(chan bool)
	go func() {
		for csvPlace := range placesChan {

			place := places.Place{}

			place.ID = csvPlace.ID
			place.Class = places.Class(csvPlace.ID >> 32)
			place.Type = csvPlace.Type
			place.Name = csvPlace.Name
			if place.Class == places.LocationClass || place.Class == places.HouseNumberClass {
				street, exists := placeMap[csvPlace.StreetID]
				if !exists {
					panic(fmt.Errorf("a street (place) with the id '%d' does not yet exist", csvPlace.StreetID))
				}
				place.Street = street
			}
			place.HouseNumber = csvPlace.HouseNumber
			district, exists := districtsMap[csvPlace.Postcode]
			if !exists {
				panic(fmt.Errorf("a district (postcode) with the id '%s' does not exist", csvPlace.Postcode))
			}
			place.District = district
			if place.Class == places.StreetClass && csvPlace.Length > 0 {
				place.Length = csvPlace.Length
			}
			place.Lat = csvPlace.Lat
			place.Lon = csvPlace.Lon
			if place.Class == places.StreetClass || place.Class == places.LocationClass {
				simpleName := places.SanitizeString(csvPlace.Name)
				place.SimpleName = simpleName
			}
			if place.Class == places.HouseNumberClass {
				place.Street.HouseNumbers = append(place.Street.HouseNumbers, &place)
			}

			placeMap[place.ID] = &place

			counts[place.Class] += 1
		}
		placesDoneChan <- true
	}()
	if err := gocsv.UnmarshalToChan(provider.PlacesReader, placesChan); err != nil {
		return nil, nil, nil, err
	}
	<-placesDoneChan

	metrics := places.Metrics{
		StreetCount:      counts[places.StreetClass],
		LocationCount:    counts[places.LocationClass],
		HouseNumberCount: counts[places.HouseNumberClass],
	}
	return districtsMap, placeMap, &metrics, nil
}
