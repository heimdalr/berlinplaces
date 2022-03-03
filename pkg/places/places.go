package places

import (
	"context"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/gocarina/gocsv"
	"io"
	"sort"
	"strings"
	"sync/atomic"
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
	OSMID         string `csv:"osm_id" json:"osmID,omitempty"`
	Class         string `csv:"class" json:"class"`
	Type          string `csv:"type" json:"type"`
	Name          string `csv:"name" json:"name"`
	Street        string `csv:"street" json:"street,omitempty"`
	HouseNumber   string `csv:"house_number" json:"houseNumber,omitempty"`
	Boundary      string `csv:"boundary" json:"boundary,omitempty"`
	Neighbourhood string `csv:"neighbourhood" json:"neighbourhood,omitempty"`
	Suburb        string `csv:"suburb" json:"suburb,omitempty"`
	Postcode      string `csv:"postcode" json:"postcode"`
	City          string `csv:"city" json:"city"`
	Lat           string `csv:"lat" json:"lat"`
	Lon           string `csv:"lon" json:"lon"`
	Relevance     uint64 `json:"relevance"`
	SimpleName    string `json:"simpleName"` // the sanitized name used for lookups
}

// prefix represents precomputed completions and places for a given prefix
type prefix struct {

	// exact matches (i.e. places whose ids exactly match this prefix)
	// (there may be more than minCompletionCount exact matches)
	exact []*result

	// completions (i.e. places to suggest for this prefix (only if not at maxPrefixLength)
	completions []*result

	// places covered by this prefix (only if at maxPrefixLength)
	places []*place
}

type Places struct {

	// maximum prefix length
	maxPrefixLength int

	// the minimum number of completions to compute
	minCompletionCount int

	// the minimum input length before doing Levenshtein
	levMinimum int

	// a list of all places
	places []*place

	// a map associating places with prefixes.
	prefixMap map[string]*prefix

	// cache for longer prefixes and prefixes with typo
	cache *ristretto.Cache
}

type Metrics struct {
	PlaceCount   int
	PrefixCount  int
	CacheMetrics *ristretto.Metrics
}

func NewPlaces(csv io.Reader, maxPrefixLength, minCompletionCount, levMinimum int) (*Places, error) {

	// basic init
	places := Places{
		maxPrefixLength:    maxPrefixLength,
		minCompletionCount: minCompletionCount,
		levMinimum:         levMinimum,
	}

	// unmarshal
	err := gocsv.Unmarshal(csv, &places.places)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall CSV data: %w", err)
	}

	// compute simple names
	places.computeSimpleNames()

	// compute pm
	places.computePrefixMap()

	// initialize cache
	places.cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,     // number of keys to track frequency of (1M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return &places, nil

}

// sanitizeString to unicode letters, spaces and minus
func sanitizeString(s string) string {
	// only unicode letters, spaces and minus
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return -1
	}, s)
	// remove spaces and minus from head and tail and lower case
	return strings.ToLower(strings.Trim(s, " -"))
}

// computeSimpleNames computes the simple names.
func (bp *Places) computeSimpleNames() {
	for _, p := range bp.places {
		p.SimpleName = sanitizeString(p.Name)
	}
}

// computePrefixMap associates places with prefixes.
func (bp *Places) computePrefixMap() {

	pm := make(map[string]*prefix)

	// sort places by length and lex order
	sort.Slice(bp.places,
		func(i, j int) bool {
			li := bp.places[i].SimpleName
			lj := bp.places[j].SimpleName
			return len(li) < len(lj) || (len(li) == len(lj) && li < li)
		})

	for d := 1; d <= bp.maxPrefixLength; d++ {
		for _, p := range bp.places {
			runes := []rune(p.SimpleName)
			runesLen := len(runes)
			prefixLen := min(runesLen, d)
			remainderLength := runesLen - prefixLen
			prefixStr := string(runes[:prefixLen])

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
			if d < bp.maxPrefixLength {

				// if completions are not yet full
				if len(pm[prefixStr].completions) < bp.minCompletionCount {
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

	bp.prefixMap = pm
}

func (bp Places) Metrics() Metrics {
	return Metrics{
		PlaceCount:   len(bp.places),
		PrefixCount:  len(bp.prefixMap),
		CacheMetrics: bp.cache.Metrics,
	}
}

func (bp Places) Query(_ context.Context, input string) []*result {

	// dissect the input
	input = sanitizeString(input)
	runes := []rune(input)
	inputLength := len(runes)

	// if we have a matching cache entry return it
	cacheResults, hit := bp.cache.Get(input)
	if hit {
		if results, ok := cacheResults.([]*result); ok {
			return results
		}
	}

	// if input is longer than max prefix length
	if inputLength >= bp.maxPrefixLength {

		// compute the (max) prefix string
		prefixString := string(runes[:bp.maxPrefixLength])

		// if we have a matching entry in the prefix map
		if pf, ok := bp.prefixMap[prefixString]; ok {

			// do Levenshtein on the places associated with this prefix
			results := bp.levenshtein(pf.places, input)

			// try to cache results (i.e. we extend the prefix map by longer prefixes)
			go func() {
				_ = bp.cache.Set(input, results, 0)
			}()

			return results

		} else {

			// do Levenshtein on all places
			results := bp.levenshtein(bp.places, input)

			// try to cache results (i.e. we extend the prefix map by long "faulty" prefixes)
			go func() {
				_ = bp.cache.Set(input, results, 0)
			}()

			return results
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
		results := bp.levenshtein(bp.places, input)

		// try to cache results
		go func() {
			_ = bp.cache.Set(input, results, 0)
		}()

		return results
	}

	// as a last resort return the empty list
	return []*result{}
}

func (bp *Places) levenshtein(places []*place, text string) []*result {

	// for each place compute the Levenshtein-Distance between its simple name and the given text
	results := make([]*result, len(places))
	for i, p := range places {
		results[i] = &result{
			Distance:   levenshtein.ComputeDistance(text, p.SimpleName),
			Percentage: 0,
			Place:      p,
		}
	}

	// sort the completions slice by Levenshtein-Distance ascending and relevance descending
	sort.Slice(results, func(i, j int) bool {
		di := results[i].Distance
		dj := results[j].Distance
		return di < dj || (di == dj && results[i].Place.Relevance > results[j].Place.Relevance)
	})

	go func() {
		bp.updateRelevance(results)
	}()

	// compute the number of completions to return
	count := min(bp.minCompletionCount, len(results))

	// return the top n completions with the smallest Levenshtein-Distance
	topResults := results[:count]
	return topResults
}

func (bp *Places) updateRelevance(results []*result) {

	// for each result
	for _, r := range results {

		// if the distance is 0, the input resulted in an exact match
		if r.Distance == 0 {

			// increase relevance (thread safe)
			atomic.AddUint64(&r.Place.Relevance, 1)

		} else {

			// results are ordered by distance wrt. text, thus break as soon as not 0
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
