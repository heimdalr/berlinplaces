package data

import (
	"github.com/heimdalr/berlinplaces/pkg/places"
	"reflect"
	"strings"
	"testing"
)

const (
	DistrictsCSV = `
postcode,district
12524,Treptow-Köpenick
10961,Friedrichshain-Kreuzberg
`
	PlacesCSV = `
id,type,name,street_id,house_number,postcode,length,lat,lon
1,,Elisabeth-Feller-Weg,,,12524,10,52.51121427531362,13.433862108201659
2,,Aachener Straße,,,10961,100,52.48010401206288,13.318894891444728
3,,Aalemannufer,,,10961,1000,52.57313191552375,13.218142687594606
4294967297,restaurant,Strandlust,1,3a,12524,,52.3762307,13.657224
8589934593,,,1,1,12524,,52.4127212,13.5714066
`
)

func TestCSVProvider_Get(t *testing.T) {
	p := CSVProvider{
		DistrictsReader: strings.NewReader(DistrictsCSV),
		PlacesReader:    strings.NewReader(PlacesCSV),
	}
	districts, placesMap, counts, err := p.Get()
	if err != nil {
		t.Errorf("Got error = %v", err)
	}

	// counts
	wantDistrictCount := 2
	districtCount := len(districts)
	if districtCount != wantDistrictCount {
		t.Errorf("Got %d districts, want %d", districtCount, wantDistrictCount)
	}
	wantCounts := map[places.Class]int32{places.StreetClass: 3, places.LocationClass: 1, places.HouseNumberClass: 1}
	if !reflect.DeepEqual(counts, wantCounts) {
		t.Errorf("Got %v places, want %v", counts, wantCounts)
	}

	// house number was added to street with id 1
	if placesMap[1].HouseNumbers == nil || len(placesMap[1].HouseNumbers) != 1 {
		t.Errorf("Missing house number")
	}

}
