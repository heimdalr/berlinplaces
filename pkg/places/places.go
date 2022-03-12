package places

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/heimdalr/berlinplaces/pkg/data"
	"github.com/rs/zerolog/log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

// Places is where all happens.
type Places struct {

	// the places config
	config *Config

	// metrics collected while running
	m       sync.RWMutex
	metrics *Metrics

	// all places
	placesMap map[int]*Place

	// a list of streets and locations (needed for completion)
	streetsAndLocations []*Place

	// a map associating places with prefixes.
	prefixMap map[string]*prefix

	// cache for longer prefixes and prefixes with typo
	cache *ristretto.Cache
}

// NewPlaces initializes a new Places object.
func (c Config) NewPlaces() (*Places, error) {

	// basic init
	places := Places{config: &c, metrics: &Metrics{}}

	loaded, err := c.DataProvider.Get()
	if err != nil {
		return nil, err
	}

	// convert district list to map
	districtMap := make(map[string]*data.District)
	for _, district := range loaded.Districts {
		districtMap[district.Postcode] = district
	}

	// convert street list to place map (reassigning IDs and linking districts)
	places.placesMap = make(map[int]*Place)
	var placeID int
	var streetID2placeID = make(map[int]int)
	for _, street := range loaded.Streets {
		places.placesMap[placeID] = &Place{
			ID:         placeID,
			Class:      Street,
			Name:       street.Name,
			cluster:    street.Cluster,
			District:   districtMap[street.Postcode],
			Lat:        street.Lat,
			Lon:        street.Lon,
			Length:     street.Length,
			simpleName: sanitizeString(street.Name),
		}
		streetID2placeID[street.ID] = placeID
		placeID += 1
		places.metrics.StreetCount += 1
	}

	// extend places map by locations (assigning IDs, linking street-places and districts)
	for _, l := range loaded.Locations {
		streetPlace := places.placesMap[streetID2placeID[l.StreetID]]
		p := Place{
			ID:          placeID,
			Class:       Location,
			Type:        l.Type,
			Name:        l.Name,
			Street:      streetPlace,
			HouseNumber: l.HouseNumber,
			District:    districtMap[l.Postcode],
			Lat:         l.Lat,
			Lon:         l.Lon,
			simpleName:  sanitizeString(l.Name),
		}
		places.placesMap[placeID] = &p
		streetPlace.locations = append(streetPlace.locations, &p)
		placeID += 1
		places.metrics.LocationCount += 1
	}

	// extend places map by house numbers (assigning IDs, linking street-places and districts)
	for _, h := range loaded.HouseNumbers {
		streetPlace := places.placesMap[streetID2placeID[h.StreetID]]
		p := Place{
			ID:          placeID,
			Class:       HouseNumber,
			Street:      streetPlace,
			HouseNumber: h.HouseNumber,
			District:    districtMap[h.Postcode],
			Lat:         h.Lat,
			Lon:         h.Lon,
		}
		places.placesMap[placeID] = &p
		streetPlace.houseNumbers = append(streetPlace.houseNumbers, &p)
		placeID += 1
		places.metrics.HouseNumberCount += 1
	}

	// collect streets and locations
	sl := make([]*Place, places.metrics.StreetCount+places.metrics.LocationCount)
	i := 0
	for _, p := range places.placesMap {
		if p.Class != HouseNumber {
			sl[i] = p
			i += 1
		}
	}

	// sort streets and locations by length and then lex order
	sort.Slice(sl, func(i, j int) bool {
		return placeLesser(sl[i], sl[j])
	})
	places.streetsAndLocations = sl

	// compute placesMap
	places.computePrefixMap()
	places.metrics.PrefixCount = len(places.prefixMap)

	// initialize cache
	places.cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,     // number of keys to track frequency of (1M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		OnEvict: func(item *ristretto.Item) {
			log.Debug().Msgf("evicting %v", item.Value)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return &places, nil
}

// Config is the configuration for Places.
type Config struct {

	// MaxPrefixLength is the maximum prefixes length to precompute completions for.
	MaxPrefixLength int `json:"maxPrefixLength"`

	// MinCompletionCount is the number of completions to return.
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

	// The data provider.
	DataProvider data.Provider
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
	StreetCount      int           `json:"streetCount"`
	LocationCount    int           `json:"locationCount"`
	HouseNumberCount int           `json:"houseNumberCount"`
	PrefixCount      int           `json:"prefixCount"`
	QueryCount       int64         `json:"queryCount"`
	AvgLookupTime    time.Duration `json:"avgLookupTime"`
}

// Place represents a single place. A place may be a street, a location, or a house number / building).
type Place struct {
	ID           int
	Class        Class
	Type         string
	Name         string
	cluster      string
	Street       *Place // in case of a location or a house number, this links (up) to the street
	HouseNumber  string
	District     *data.District // this links to the postcode and district
	Lat          float64
	Lon          float64
	Length       int
	Relevance    uint64
	simpleName   string
	houseNumbers []*Place // in case of a street, this links (down) to associated house numbers
	locations    []*Place // in case of a street, this links (down) to associated locations
}

// MarshalJSON implements the JSON marshaller interface for Places.
func (p *Place) MarshalJSON() ([]byte, error) {

	// depending on the class we want to render the place slightly in a different way
	var (
		street, name     string
		streetID, length *int
	)
	switch p.Class {
	case Street:
		name = p.Name
		length = &p.Length
	case Location:
		name = p.Name
		street = p.Street.Name
		streetID = &p.Street.ID
	default: // HouseNumber
		street = p.Street.Name
		streetID = &p.Street.ID
	}
	return json.Marshal(&struct {
		ID          int     `json:"id"`
		Class       string  `json:"class"`
		Type        string  `json:"type,omitempty"`
		Name        string  `json:"name"`
		Street      string  `json:"street,omitempty"`
		StreetID    *int    `json:"streetID,omitempty"`
		HouseNumber string  `json:"houseNumber,omitempty"`
		Postcode    string  `json:"postcode"`
		District    string  `json:"district"`
		Length      *int    `json:"length,omitempty"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Relevance   uint64  `json:"relevance"`
	}{
		ID:          p.ID,
		Class:       p.Class.String(),
		Type:        p.Type,
		Name:        name,
		Street:      street,
		StreetID:    streetID,
		HouseNumber: p.HouseNumber,
		Postcode:    p.District.Postcode,
		District:    p.District.District,
		Length:      length,
		Lat:         p.Lat,
		Lon:         p.Lon,
		Relevance:   p.Relevance,
	})
}

// Result wraps a place.
type Result struct {
	Distance int    `json:"distance"`
	Place    *Place `json:"place"`
}

// Class enumerates different place types / classes.
type Class int

const (

	// Street is the place class of streets.
	Street = iota

	// Location is place class of locations.
	Location

	// HouseNumber is the place class of house numbers / buildings.
	HouseNumber
)

// String implements the stringer interface for Class.
func (c Class) String() string {
	return [...]string{"street", "location", "csvHouseNumber"}[c]
}

// prefix represents precomputed completions and places for a given prefix
type prefix struct {

	// completions (i.e. places to suggest for this prefix (only if < MaxPrefixLength)
	completions []*Result

	// places covered by this prefix (if < MaxPrefixLength those are the places in the completions)
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

// GetCompletions returns completions for the given input.
func (bp *Places) GetCompletions(ctx context.Context, input string) []*Result {
	start := time.Now()
	r := bp.getCompletions(ctx, input)
	go bp.updateMetrics(time.Since(start))
	return r
}

func (bp *Places) getCompletions(_ context.Context, input string) []*Result {

	// dissect the input
	simpleInput := sanitizeString(input)
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
		prefixString := string(runes[:min(len(runes), bp.config.MaxPrefixLength)])

		// if we have a matching entry in the prefix map
		if pf, ok := bp.prefixMap[prefixString]; ok {

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

	// if we have a matching entry in the prefixMap, then return the completions for that
	if pf, ok := bp.prefixMap[simpleInput]; ok {

		// update relevance
		go bp.updateRelevance(pf.completions, simpleInput)

		return pf.completions
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

func (bp *Places) GetPlace(ctx context.Context, placeID int, houseNumber string) *Place {
	start := time.Now()
	p := bp.getPlace(ctx, placeID, houseNumber)
	go bp.updateMetrics(time.Since(start))
	return p
}

func (bp *Places) getPlace(_ context.Context, placeID int, houseNumber string) *Place {
	if p, ok := bp.placesMap[placeID]; ok {
		if houseNumber == "" {
			return p
		} else {
			for _, h := range p.houseNumbers {
				if h.HouseNumber == houseNumber {
					return h
				}
			}
		}
	}

	return nil
}

// computePrefixMap associates prefixes with completions xor places.
func (bp *Places) computePrefixMap() {

	pm := make(map[string]*prefix)

	for d := 1; d <= bp.config.MaxPrefixLength; d++ {
		for _, p := range bp.streetsAndLocations {
			runes := []rune(p.simpleName)
			runesLen := len(runes)
			prefixLen := min(runesLen, d)
			remainderLength := runesLen - prefixLen
			prefixStr := string(runes[:prefixLen])

			// as we are here we have something for this prefix, init a map entry (if necessary)
			if _, ok := pm[prefixStr]; !ok {
				pm[prefixStr] = &prefix{}
			}

			// append this place as a completion and place if below MaxPrefixLength and
			// - its id exactly matches the current prefix or
			// - we don't have enough completions yet
			if d < bp.config.MaxPrefixLength {
				if remainderLength == 0 || len(pm[prefixStr].completions) < bp.config.MinCompletionCount {
					r := Result{
						Distance: remainderLength,
						Place:    p,
					}
					pm[prefixStr].places = append(pm[prefixStr].places, p)
					pm[prefixStr].completions = append(pm[prefixStr].completions, &r)
					continue
				}
			} else {

				// we are at or above MaxPrefixLength thus at the place as place
				pm[prefixStr].places = append(pm[prefixStr].places, p)
			}
		}

		bp.prefixMap = pm
	}
}

// updateRelevance increases the relevance for each exact match in the results
// slice and returns a slice containing the updated elements (if any).
func (bp *Places) updateRelevance(results []*Result, simpleInput string) []*Place {

	var updatedPlaces []*Place

	// for each result
	for _, r := range results {

		// if the particular result in an exact match
		if r.Place.simpleName == simpleInput {

			// increase relevance (thread safe)
			atomic.AddUint64(&r.Place.Relevance, 1)

			updatedPlaces = append(updatedPlaces, r.Place)
		}
	}

	// update prefix completions if needed
	if len(updatedPlaces) > 0 {
		bp.updateCompletions(updatedPlaces)
	}

	return updatedPlaces
}

// updateCompletions updates completions for the given places (which must have
// all the same simpleName - see updateRelevance).
func (bp *Places) updateCompletions(updatedPlaces []*Place) {

	simpleName := updatedPlaces[0].simpleName
	runes := []rune(simpleName)
	runesLen := len(runes)

	// assertion about updated place names
	for _, p := range updatedPlaces {
		if p.simpleName != simpleName {
			panic("unexpected place name")
		}
	}

	for d := 1; d < min(bp.config.MaxPrefixLength, runesLen); d++ {

		prefixStr := string(runes[:d])

		// get the current completions for this prefix
		currentPlaces := bp.prefixMap[prefixStr].places

		// merge the results that where updated to the current completions and deduplicate
		mergedPlaces := deDuplicate(append(currentPlaces, updatedPlaces...))

		// do Levenshtein on the merged places wrt. the prefix string
		results := bp.levenshtein(mergedPlaces, prefixStr)

		var newCompletions []*Result
		var newPlaces []*Place
		for _, r := range results {
			if r.Place.simpleName == prefixStr || len(newCompletions) < bp.config.MinCompletionCount {
				newCompletions = append(newCompletions, r)
				newPlaces = append(newPlaces, r.Place)
			}
		}

		bp.prefixMap[prefixStr].places = newPlaces
		bp.prefixMap[prefixStr].completions = newCompletions

	}
}

func (bp *Places) levenshtein(places []*Place, simpleInput string) []*Result {

	// for each place compute the Levenshtein-Distance between its simple name and the given simple input
	results := make([]*Result, len(places))
	for i, p := range places {
		results[i] = &Result{
			Distance: levenshtein.ComputeDistance(simpleInput, p.simpleName),
			Place:    p,
		}
	}

	// sort results via place ranking.
	sort.Slice(results, func(i, j int) bool {
		return bp.resultRanking(results[i], results[j])
	})

	// compute the number of completions to return (i.e. all exact matches filled up to MinCompletionCount)
	count := min(bp.config.MinCompletionCount, len(results))
	for i := count; i < len(results); i++ {
		if results[i].Place.simpleName == simpleInput {

			// we are past MinCompletionCount but still have an exact match, therefore add it
			count += 1
		} else {

			// as results are ordered by distance, we can break here
			break
		}
	}

	// return the completions
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
	distanceDelta := abs(di - dj)
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
		if pi.Class == Street {
			return true
		} else {
			return false
		}
	}

	// if streets, less by longer length (by the above clause, types must be equal)
	if pi.Class == Street {
		if pi.Length > pj.Length {
			return true
		} else {
			return false
		}
	}

	return false
}

// placeLesser compares two places wrt. string length and lexical order.
// placeLesser return true, if the first place is less than the second one.
func placeLesser(i, j *Place) bool {

	// less by character length
	if len(i.simpleName) != len(j.simpleName) {
		if len(i.simpleName) < len(j.simpleName) {
			return true
		} else {
			return false
		}
	}

	// less by lex
	if i.simpleName != j.simpleName {
		if i.simpleName < j.simpleName {
			return true
		} else {
			return false
		}
	}

	// defaults
	return false
}

// deDuplicate removes duplicate places.
func deDuplicate(results []*Place) []*Place {

	ids := make(map[int]interface{})
	var places []*Place

	for _, p := range results {
		if _, exists := ids[p.ID]; !exists {
			ids[p.ID] = struct{}{}
			places = append(places, p)
		}
	}

	return places
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

// min returns the minimum of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
