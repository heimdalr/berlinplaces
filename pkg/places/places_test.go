package places_test

import (
	"context"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/spf13/viper"
	"strings"
	"testing"
)

func TestPlaces_query(t *testing.T) {
	districtsCSV := `
postcode,csvDistrict
12524,Treptow-Köpenick
10961,Friedrichshain-Kreuzberg
`
	streetsCSV := `
id,name,cluster,postcode,lat,lon,length
1,Elisabeth-Feller-Weg,1,12524,52.51121427531362,13.433862108201659, 10
2,Aachener Straße,1,10961,52.48010401206288,13.318894891444728, 100
3,Aalemannufer,1,10961,52.57313191552375,13.218142687594606, 1000
`

	locationsCSV := `
type,name,street_id,housenumber,postcode,lat,lon
restaurant,Strandlust,1,3a,12527,52.3762307,13.657224
`
	housenumbersCSV := `
street_id,housenumber,postcode,lat,lon
1,1,12524,52.4127212,13.5714066
`
	viper.SetDefault("MAX_PREFIX_LENGTH", 4)
	viper.SetDefault("MIN_COMPLETION_COUNT", 10)
	viper.SetDefault("LEV_MINIMUM", 4)

	p, err := places.NewPlaces(
		strings.NewReader(districtsCSV),
		strings.NewReader(streetsCSV),
		strings.NewReader(locationsCSV),
		strings.NewReader(housenumbersCSV),
	)

	if err != nil {
		t.Fatal(err)
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
