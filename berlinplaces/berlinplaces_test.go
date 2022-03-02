package berlinplaces

import (
	"context"
	"testing"
)

func TestPlaces_simple(t *testing.T) {
	places := []*place{
		{
			Name:     "Elisabeth-Feller-Weg",
			Postcode: "12205",
		},
		{
			Name:     "Krokusstraße",
			Postcode: "12357",
		},
	}
	pm := computePrefixMap(places, 3, 3)

	berlinPlaces := &BerlinPlaces{
		places:             places,
		prefixMap:          pm,
		maxPrefixLength:    3,
		minCompletionCount: 3,
		levMinimum:         2,
	}
	r := berlinPlaces.Query(context.Background(), "Elisabeth-Feller-Weg")
	if len(r) != 1 {
		t.Errorf("expected 1, got %d", len(r))
	}
	got := r[0].Place.Name
	want := "Elisabeth-Feller-Weg"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

}
func TestPlaces_query(t *testing.T) {
	berlinPlaces, err := NewBerlinPlaces("berlin.csv", 8, 5, 4)
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
