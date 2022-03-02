package places_test

import (
	"context"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"strings"
	"testing"
)

func TestPlaces_query(t *testing.T) {
	csvString := `
place_id,parent_place_id,class,type,name,street,housenumber,suburb,postcode,city,lat,lon
541950,465999,highway,living_street,Elisabeth-Feller-Weg,,,,12205,,52.4280125,13.2992439
621178,709865,highway,residential,Krokusstraße,,,,12357,,52.4229373,13.4951325
`
	csvReader := strings.NewReader(csvString)
	berlinPlaces, err := places.NewPlaces(csvReader, 8, 5, 4)
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
			r := berlinPlaces.Query(context.Background(), tt.text)
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
