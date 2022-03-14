package places_test

import (
	"context"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/heimdalr/berlinplaces/pkg/data"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"strings"
	"testing"
)

type testProvider struct{}

func (_ testProvider) Get() (*data.Data, error) {

	districtsCSV := `
postcode,csvDistrict
12524,Treptow-Köpenick
10961,Friedrichshain-Kreuzberg
`
	streetsCSV := `
id,name,postcode,lat,lon,length
s1,Elisabeth-Feller-Weg,12524,52.51121427531362,13.433862108201659, 10
s2,Aachener Straße,10961,52.48010401206288,13.318894891444728, 100
s3,Aalemannufer,10961,52.57313191552375,13.218142687594606, 1000
`

	locationsCSV := `
id,type,name,street_id,housenumber,postcode,lat,lon
l1,restaurant,Strandlust,s1,3a,12527,52.3762307,13.657224
`
	housenumbersCSV := `
id,street_id,housenumber,postcode,lat,lon
h1,s1,1,12524,52.4127212,13.5714066
`

	gocsv.SetHeaderNormalizer(func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", "")
	})

	d := data.Data{}

	jobs := []struct {
		csvData string
		data    interface{}
	}{
		{districtsCSV, &d.Districts},
		{streetsCSV, &d.Streets},
		{locationsCSV, &d.Locations},
		{housenumbersCSV, &d.HouseNumbers},
	}

	for _, j := range jobs {

		// unmarshall into given interface
		err := gocsv.Unmarshal(strings.NewReader(j.csvData), j.data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall data: %w", err)
		}

	}

	return &d, nil
}

func TestPlaces_GetCompletions(t *testing.T) {

	p, err := places.DefaultConfig.NewPlaces(testProvider{})
	if err != nil {
		t.Fatal(fmt.Errorf("failed to init places: %w", err))
	}

	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Elisabeth-Feller-Weg (Exact)",
			text: "Elisabeth-Feller-Weg",
			want: "Elisabeth-Feller-Weg",
		},
		{
			name: "ElisabFeller-Weg",
			text: "ElisabFeller-Weg",
			want: "Elisabeth-Feller-Weg",
		},
		{
			name: "Eisabeth-Feller-Weg (Typo Beginning)",
			text: "Eisabeth-Feller-Weg",
			want: "Elisabeth-Feller-Weg",
		},
		{
			name: "Elisabeth-Felle-Weg (Typo End)",
			text: "Elisabeth-Felle-Weg",
			want: "Elisabeth-Feller-Weg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := p.GetCompletions(context.Background(), tt.text)
			rLength := len(r)
			if rLength < 1 {
				t.Errorf("got %d, want > 0", rLength)
			}
			got := r[0].Place.Name
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
