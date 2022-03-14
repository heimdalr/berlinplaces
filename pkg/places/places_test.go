package places_test

import (
	"context"
	"fmt"
	"github.com/heimdalr/berlinplaces/pkg/data"
	"github.com/heimdalr/berlinplaces/pkg/places"
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

func TestPlaces_GetCompletions(t *testing.T) {

	dataProvider := data.CSVProvider{
		DistrictsReader: strings.NewReader(DistrictsCSV),
		PlacesReader:    strings.NewReader(PlacesCSV),
	}

	p, err := places.DefaultConfig.NewPlaces(dataProvider)
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
