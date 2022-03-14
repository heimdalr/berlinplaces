package places

import (
	"context"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/dgraph-io/ristretto"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Config is the configuration for Places.
type Config struct {

	// MaxPrefixLength is the maximum prefixes length to precompute results for.
	MaxPrefixLength int `json:"maxPrefixLength"`

	// MinCompletionCount is the number of results to return.
	MinCompletionCount int `json:"minCompletionCount"`

	// MinLev is the minimum input length before doing Levenshtein comparison.
	MinLev int `json:"minLev"`

	// DistanceCut is used in result ranking. DistanceCut is the delta in distances
	// to ignore in favor of relevance (unless one of the results has a distance of
	// 0).
	DistanceCut int `json:"distanceCut"`

	// Duration to wait before evicting cache entries (in order to consider
	// potentially changed relevance values).
	CacheTTL time.Duration `json:"cacheTTL"`
}

// DefaultConfig is the default configuration for Places.
var DefaultConfig = &Config{
	MaxPrefixLength:    4,
	MinCompletionCount: 5,
	MinLev:             4,
	DistanceCut:        4,
	CacheTTL:           300 * time.Second,
}

// Metrics is the type to sore metrics.
type Metrics struct {
	StreetCount      int32         `json:"streetCount"`
	LocationCount    int32         `json:"locationCount"`
	HouseNumberCount int32         `json:"houseNumberCount"`
	PrefixCount      int           `json:"prefixCount"`
	QueryCount       int64         `json:"queryCount"`
	AvgLookupTime    time.Duration `json:"avgLookupTime"`
}

// Places is where all happens.
type Places struct {

	// the places config
	config *Config

	// metrics collected while running
	m       sync.RWMutex
	metrics *Metrics

	// districts mapped by postcode
	districtsMap map[string]*District

	// places mapped by place ID
	placesMap map[int64]*Place

	// a sorted slice of street- and location places (needed for completion-computation)
	streetsAndLocations []*Place

	// precomputed completions mapped by a prefix
	prefixCompletions map[string]*completion

	// cache for longer prefixes and prefixes with typo
	cache *ristretto.Cache
}

type Provider interface {
	Get() (DistrictMap, PlaceMap, *Metrics, error)
}

// NewPlaces initializes a new Places object.
func (config Config) NewPlaces(dataProvider Provider) (*Places, error) {

	// get districts and places
	if dataProvider == nil {
		return nil, fmt.Errorf("data provider must not be nil")
	}
	districtsMap, placesMap, metrics, errData := dataProvider.Get()
	if errData != nil {
		return nil, errData
	}

	// collect streets and locations into a slice and then sort it
	streetsAndLocations := make([]*Place, metrics.StreetCount+metrics.LocationCount)
	i := 0
	for _, place := range placesMap {
		if place.Class == StreetClass || place.Class == LocationClass {
			streetsAndLocations[i] = place
			i += 1
		}
	}
	sort.Slice(streetsAndLocations, func(i, j int) bool {
		return placeLesser(streetsAndLocations[i], streetsAndLocations[j])
	})

	// compute prefix completions
	prefixCompletions := computePrefixCompletions(streetsAndLocations, config.MaxPrefixLength, config.MaxPrefixLength)
	metrics.PrefixCount = len(prefixCompletions)

	// initialize cache
	cache, errCache := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,     // number of keys to track frequency of (1M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if errCache != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", errCache)
	}

	// basic init
	//places := Places{config: &config, metrics: &Metrics{}}
	return &Places{
		config:              &config,
		m:                   sync.RWMutex{},
		metrics:             metrics,
		districtsMap:        districtsMap,
		placesMap:           placesMap,
		streetsAndLocations: streetsAndLocations,
		prefixCompletions:   prefixCompletions,
		cache:               cache,
	}, nil

}

// Result wraps a place.
type Result struct {
	Distance int    `json:"distance"`
	Place    *Place `json:"place"`
}

// completion represents precomputed results and places (for a given prefix)
type completion struct {

	// results (i.e. places to suggest for this prefix (only if < MaxPrefixLength)
	results []*Result

	// places covered by this prefix (if < MaxPrefixLength those are the places in the results)
	places []*Place
}

// Config returns the configuration.
func (bp *Places) Config() Config {
	return *bp.config
}

// Metrics returns current metrics.
func (bp *Places) Metrics() Metrics {
	return *bp.metrics
}

// GetCompletions returns results for the given input.
func (bp *Places) GetCompletions(ctx context.Context, input string) []*Result {
	start := time.Now()
	r := bp.getCompletions(ctx, input)
	go bp.updateMetrics(time.Since(start))
	return r
}

func (bp *Places) getCompletions(_ context.Context, input string) []*Result {

	// dissect the input
	simpleInput := SanitizeString(input)
	runes := []rune(simpleInput)
	inputLength := len(runes)

	// if we have a matching cache entry return it
	cacheResults, hit := bp.cache.Get(simpleInput)
	if hit {
		if results, ok := cacheResults.([]*Result); ok {

			// update relevance
			go bp.updateRelevance(results, simpleInput)

			return results
		} else {
			panic("failed to cast cache results")
		}
	}

	// if simpleInput is longer or equal to than MaxPrefixLength
	if inputLength >= bp.config.MaxPrefixLength {

		// compute the (max) prefix string
		prefixString := string(runes[:Min(len(runes), bp.config.MaxPrefixLength)])

		// if we have a matching entry in the prefix map
		if pf, ok := bp.prefixCompletions[prefixString]; ok {

			// do Levenshtein on the places associated with this prefix
			results := bp.levenshtein(pf.places, simpleInput)

			go func() {

				// update relevance
				bp.updateRelevance(results, simpleInput)

				// try to cache results (i.e. we extend the prefix map by longer prefixes)
				bp.cache.SetWithTTL(simpleInput, results, 0, bp.config.CacheTTL)
			}()

			return results

		} else {

			// do Levenshtein on all streets and locations
			results := bp.levenshtein(bp.streetsAndLocations, simpleInput)

			go func() {

				// update relevance
				bp.updateRelevance(results, simpleInput)

				// try to cache results (i.e. we extend the prefix map by long "faulty" prefixes)
				bp.cache.SetWithTTL(simpleInput, results, 0, bp.config.CacheTTL)
			}()

			return results
		}
	}

	// simpleInput length is smaller than max prefix length thus the simpleInput is the prefix to match

	// if we have a matching entry in the prefixCompletions, then return the results for that
	if pf, ok := bp.prefixCompletions[simpleInput]; ok {

		// update relevance
		go bp.updateRelevance(pf.results, simpleInput)

		return pf.results
	}

	// there is no matching prefix, but above MinLev
	if inputLength >= bp.config.MinLev {

		// do levenshtein on all streets and location
		results := bp.levenshtein(bp.streetsAndLocations, simpleInput)

		go func() {

			// update relevance for exact matches
			bp.updateRelevance(results, simpleInput)

			// try to cache results
			bp.cache.SetWithTTL(simpleInput, results, 0, bp.config.CacheTTL)
		}()

		return results
	}

	// as a last resort return the empty list
	return []*Result{}
}

func (bp *Places) GetPlace(ctx context.Context, placeID int64, houseNumber string) *Place {
	start := time.Now()
	p := bp.getPlace(ctx, placeID, houseNumber)
	go bp.updateMetrics(time.Since(start))
	return p
}

func (bp *Places) getPlace(_ context.Context, placeID int64, houseNumber string) *Place {
	if p, ok := bp.placesMap[placeID]; ok {
		if houseNumber == "" {
			return p
		} else {
			if p.Class == StreetClass && p.HouseNumbers != nil {
				for _, house := range p.HouseNumbers {
					if house.HouseNumber == houseNumber {
						return house
					}
				}
			}
		}
	}

	return nil
}

// computePrefixCompletions compute completions for prefixes of street- and location names.
func computePrefixCompletions(streetsAndLocations []*Place, maxPrefixLength, minCompletionCount int) map[string]*completion {

	pc := make(map[string]*completion)

	for d := 1; d <= maxPrefixLength; d++ {
		for _, p := range streetsAndLocations {
			runes := []rune(p.SimpleName)
			runesLen := len(runes)
			prefixLen := Min(runesLen, d)
			remainderLength := runesLen - prefixLen
			prefixStr := string(runes[:prefixLen])

			// as we are here we have something for this prefix, init a map entry (if necessary)
			if _, ok := pc[prefixStr]; !ok {
				pc[prefixStr] = &completion{}
			}

			// append this place as a completion and place if below MaxPrefixLength and
			// - its id exactly matches the current prefix or
			// - we don't have enough results yet
			if d < maxPrefixLength {

				if remainderLength == 0 || len(pc[prefixStr].results) < minCompletionCount {
					r := Result{
						Distance: remainderLength,
						Place:    p,
					}
					pc[prefixStr].places = append(pc[prefixStr].places, p)
					pc[prefixStr].results = append(pc[prefixStr].results, &r)
					continue
				}
			} else {

				// we are at or above MaxPrefixLength thus at the place as place
				pc[prefixStr].places = append(pc[prefixStr].places, p)
			}
		}
	}
	return pc
}

// updateRelevance increases the relevance for each exact match in the results
// slice and returns a slice containing the updated elements (if any).
func (bp *Places) updateRelevance(results []*Result, simpleInput string) []*Place {

	var updatedPlaces []*Place

	// for each result
	for _, r := range results {

		// if the particular result in an exact match
		if r.Place.SimpleName == simpleInput {

			// increase relevance (thread safe)
			atomic.AddUint64(&r.Place.Relevance, 1)

			updatedPlaces = append(updatedPlaces, r.Place)
		}
	}

	// update prefix results if needed
	if len(updatedPlaces) > 0 {
		bp.updateCompletions(updatedPlaces)
	}

	return updatedPlaces
}

// updateCompletions updates results for the given places (which must have
// all the same simpleName - see updateRelevance).
func (bp *Places) updateCompletions(updatedPlaces []*Place) {

	simpleName := updatedPlaces[0].SimpleName
	runes := []rune(simpleName)
	runesLen := len(runes)

	// assertion about updated place names
	for _, p := range updatedPlaces {
		if p.SimpleName != simpleName {
			panic("unexpected place name")
		}
	}

	for d := 1; d < Min(bp.config.MaxPrefixLength, runesLen); d++ {

		prefixStr := string(runes[:d])

		// get the current results for this prefix
		currentPlaces := bp.prefixCompletions[prefixStr].places

		// merge the results that where updated to the current results and deduplicate
		mergedPlaces := deDuplicate(append(currentPlaces, updatedPlaces...))

		// do Levenshtein on the merged places wrt. the prefix string
		results := bp.levenshtein(mergedPlaces, prefixStr)

		var newCompletions []*Result
		var newPlaces []*Place
		for _, r := range results {
			if r.Place.SimpleName == prefixStr || len(newCompletions) < bp.config.MinCompletionCount {
				newCompletions = append(newCompletions, r)
				newPlaces = append(newPlaces, r.Place)
			}
		}

		bp.prefixCompletions[prefixStr].places = newPlaces
		bp.prefixCompletions[prefixStr].results = newCompletions

	}
}

func (bp *Places) levenshtein(places []*Place, simpleInput string) []*Result {

	// for each place compute the Levenshtein-Distance between its simple name and the given simple input
	results := make([]*Result, len(places))
	for i, p := range places {
		results[i] = &Result{
			Distance: levenshtein.ComputeDistance(simpleInput, p.SimpleName),
			Place:    p,
		}
	}

	// sort results via place ranking.
	sort.Slice(results, func(i, j int) bool {
		return bp.resultRanking(results[i], results[j])
	})

	// compute the number of results to return (i.e. all exact matches filled up to MinCompletionCount)
	count := Min(bp.config.MinCompletionCount, len(results))
	for i := count; i < len(results); i++ {
		if results[i].Place.SimpleName == simpleInput {

			// we are past MinCompletionCount but still have an exact match, therefore add it
			count += 1
		} else {

			// as results are ordered by distance, we can break here
			break
		}
	}

	// return the results
	return results[:count]
}

func (bp *Places) updateMetrics(duration time.Duration) {
	bp.m.Lock()
	defer bp.m.Unlock()
	bp.metrics.QueryCount += 1
	if bp.metrics.AvgLookupTime == 0 {
		bp.metrics.AvgLookupTime = duration
	} else {
		bp.metrics.AvgLookupTime = (bp.metrics.AvgLookupTime + duration) / 2
	}
}

// resultRanking compares two levenshtein results wrt. distance, relevance, class, and (in case of streets) length.
// resultRanking returns true, if the first place should be ranked higher than the second one. resultRanking should
// be used in sorting slices (analog to the lesser function) sorting higher ranks to the beginning.
func (bp *Places) resultRanking(i, j *Result) bool {

	di := i.Distance
	dj := j.Distance

	// If one of the distances is 0 (i.e. an exact match) but not both rank the exact match higher.
	if di != dj {
		if di == 0 {
			return true
		}
		if dj == 0 {
			return false
		}
	}

	// If none of the results is an exact match but the delta in distances is greater than DistanceCut
	// the one with the smaller distance will be ranked higher.
	distanceDelta := Abs(di - dj)
	if distanceDelta > bp.config.DistanceCut {
		if di < dj {
			return true
		} else {
			return false
		}
	}

	pi := i.Place
	pj := j.Place

	// As there is no exact match and the delta in distances is within DistanceCut,
	// rank by relevance (if different).
	if pi.Relevance != pj.Relevance {
		if pi.Relevance > pj.Relevance {
			return true
		} else {
			return false
		}
	}

	// relevance is equal, come back to distances
	if di != dj {
		if di < dj {
			return true
		} else {
			return false
		}
	}

	// results have same distance and relevance, rank streets over locations
	if pi.Class != pj.Class {
		if pi.Class == StreetClass {
			return true
		} else {
			return false
		}
	}

	// if streets, less by longer length (by the above clause, types must be equal)
	if pi.Class == StreetClass {
		if pi.Length > pj.Length {
			return true
		} else {
			return false
		}
	}

	return false
}
