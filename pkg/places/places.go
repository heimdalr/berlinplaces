package places

import (
	"context"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/gocarina/gocsv"
	"io"
	"sort"
	"strings"
	"unicode"
)

type result struct {
	Distance   int    `json:"distance"`
	Percentage int    `json:"percentage"`
	Place      *place `json:"place"`
}

type place struct {
	PlaceID       string `csv:"place_id" json:"placeID"`
	ParentPlaceID string `csv:"parent_place_id" json:"parentPlaceID"`
	OSMID         string `csv:"osm_id" json:"osmID"`
	Class         string `csv:"class" json:"class"`
	Type          string `csv:"type" json:"type"`
	Name          string `csv:"name" json:"name"`
	Street        string `csv:"street" json:"street"`
	HouseNumber   string `csv:"house_number" json:"houseNumber"`
	Suburb        string `csv:"suburb" json:"suburb"`
	Postcode      string `csv:"postcode" json:"postcode"`
	City          string `csv:"city" json:"city"`
	Lat           string `csv:"lat" json:"lat"`
	Lon           string `csv:"lon" json:"lon"`
}

func (p place) id() string {
	return sanitizeString(p.Name)
}

// prefix represents precomputed completions and places for a given prefix
type prefix struct {

	// exact matches (i.e. places whose ids exactly match this prefix)
	// (there may be more than minCompletionCount exact matches)
	exact []*result

	// completions (i.e. places to suggest for this prefix (only if not isMaxDepth)
	completions []*result

	// places covered by this prefix (only if isMaxDepth)
	places []*place
}

type Places struct {

	// a list of all places
	places []*place

	// maximum prefix length
	maxPrefixLength int

	// a map associating places with prefixes.
	prefixMap map[string]*prefix

	// the minimum number of completions to compute
	minCompletionCount int

	// the minimum input length before doing Levenshtein
	levMinimum int
}

func NewPlaces(csv io.Reader, maxPrefixLength, minCompletionCount, levMinimum int) (*Places, error) {

	// unmarshal
	var places []*place
	if err := gocsv.Unmarshal(csv, &places); err != nil {
		return nil, fmt.Errorf("failed to unmarshall CSV data: %w", err)
	}

	// compute pm
	pm := computePrefixMap(places, maxPrefixLength, minCompletionCount)

	return &Places{
		places:             places,
		prefixMap:          pm,
		maxPrefixLength:    maxPrefixLength,
		minCompletionCount: minCompletionCount,
		levMinimum:         levMinimum,
	}, nil
}

// sanitizeString to unicode letters, spaces and minus
func sanitizeString(s string) string {
	// only unicode letters, spaces and minus
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || r == 32 || r == 45 {
			return r
		}
		return -1
	}, s)
	// remove spaces and minus from head and tail
	return strings.Trim(s, " -")
}

// computePrefixMap associates places with prefixes.
func computePrefixMap(allPlaces []*place, maxPrefixLength, minCompletionCount int) map[string]*prefix {

	pm := make(map[string]*prefix)

	// sort allPlaces by name
	sort.Slice(allPlaces,
		func(i, j int) bool {
			li := allPlaces[i].Name
			lj := allPlaces[j].Name
			return len(li) < len(lj) || (len(li) == len(lj) && allPlaces[i].Name < allPlaces[j].Name)
		})

	for d := 1; d <= maxPrefixLength; d++ {
		for _, p := range allPlaces {
			id := p.id()
			runes := []rune(id)
			runesLen := len(runes)
			prefixLen := min(runesLen, d)
			remainderLength := runesLen - prefixLen
			prefixStr := strings.ToLower(string(runes[:prefixLen]))

			// as we are here we have something for this prefix, init a map entry (if necessary)
			if _, ok := pm[prefixStr]; !ok {
				pm[prefixStr] = &prefix{}
			}

			// append this place as an exact match, if its id exactly matches the current prefix
			if remainderLength == 0 {

				r := result{
					Distance:   0,
					Percentage: 0,
					Place:      p,
				}
				pm[prefixStr].exact = append(pm[prefixStr].exact, &r)
				continue
			}

			// if not at max maxPrefixLength
			if d < maxPrefixLength {

				// if completions are not yet full
				if len(pm[prefixStr].completions) < minCompletionCount {
					r := result{
						Distance:   remainderLength,
						Percentage: 0,
						Place:      p,
					}

					// add place as completion
					pm[prefixStr].completions = append(pm[prefixStr].completions, &r)
				}

			} else {

				// we are at max maxPrefixLength so collect this place
				pm[prefixStr].places = append(pm[prefixStr].places, p)
			}
		}
	}

	return pm
}

func (bp Places) Query(_ context.Context, input string) []*result {

	// dissect the input
	input = strings.ToLower(sanitizeString(input))
	runes := []rune(input)
	inputLength := len(runes)

	// if input is longer than max prefix length
	if inputLength >= bp.maxPrefixLength {

		// compute the (max) prefix string
		prefixString := string(runes[:bp.maxPrefixLength])

		// if we have a matching entry in the prefix map
		if pf, ok := bp.prefixMap[prefixString]; ok {

			// do Levenshtein on the places associated with this prefix
			return bp.levenshtein(pf.places, input)
		} else {

			// do Levenshtein on all places
			return bp.levenshtein(bp.places, input)
		}
	}

	// input length is smaller than max prefix length thus the input is the prefix to match

	// if we have a matching entry in the prefixMap, then there must be exact matches and / or completions
	if pf, ok := bp.prefixMap[input]; ok {

		// if there are more exact matches than needed, return them (all)
		if len(pf.exact) >= bp.minCompletionCount {
			return pf.exact
		}

		// combine exact matches and completions and return the those
		combined := append(pf.exact, pf.completions...)
		return combined
	}

	// there is no matching prefix, but above levMinimum
	if inputLength >= bp.levMinimum {

		// do levenshtein on the complete set of places
		return bp.levenshtein(bp.places, input)
	}

	// as a last resort return the empty list
	return []*result{}
}

func (bp Places) levenshtein(places []*place, text string) []*result {

	// for each place compute the Levenshtein-Distance between its name and the given text
	results := make([]*result, len(places))
	for i, p := range places {
		results[i] = &result{
			Distance:   levenshtein.ComputeDistance(text, p.Name),
			Percentage: 0,
			Place:      p,
		}
	}

	// sort the completions slice by Levenshtein-Distance in ascending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	// compute the number of completions to return
	count := min(bp.minCompletionCount, len(results))

	// return the top n completions with the smallest Levenshtein-Distance
	topResults := results[:count]
	return topResults
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
