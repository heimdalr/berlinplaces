package places

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

type district struct {
	Postcode string `csv:"postcode"`
	District string `csv:"district"`
}

type street struct {
	ID       int     `csv:"id"`
	Name     string  `csv:"name"`
	Cluster  string  `csv:"cluster"`
	Postcode string  `csv:"postcode"`
	Lat      float64 `csv:"lat"`
	Lon      float64 `csv:"lon"`
	Length   int32   `csv:"length"`
}

type location struct {
	Type        string  `csv:"type"`
	Name        string  `csv:"name"`
	StreetID    int     `csv:"street_id"`
	HouseNumber string  `csv:"house_number"`
	Postcode    string  `csv:"postcode"`
	Lat         float64 `csv:"lat"`
	Lon         float64 `csv:"lon"`
}

type houseNumber struct {
	StreetID    int     `csv:"street_id"`
	HouseNumber string  `csv:"house_number"`
	Postcode    string  `csv:"postcode"`
	Lat         float64 `csv:"lat"`
	Lon         float64 `csv:"lon"`
}

type PlaceClass int

const (
	streetClass = iota
	locationClass
	houseNumberClass
)

func (pc PlaceClass) String() string {
	return [...]string{"street", "location", "houseNumber"}[pc]
}

func (pc *PlaceClass) MarshalJSON() ([]byte, error) {
	return json.Marshal(pc.String())
}

type place struct {
	ID           int
	Class        PlaceClass
	Type         string
	Name         string
	cluster      string
	Street       *place // in case of a location or a house number, this links (up) to the street
	HouseNumber  string
	District     *district // this links to the postcode and district
	Lat          float64
	Lon          float64
	Length       int32
	Relevance    uint64
	simpleName   string
	houseNumbers []*place // in case of a street, this links (down) to associated house numbers
	locations    []*place // in case of a street, this links (down) to associated locations
}

func (p *place) MarshalJSON() ([]byte, error) {
	if p.Class == streetClass {
		return json.Marshal(&struct {
			ID        int        `json:"id"`
			Class     PlaceClass `json:"class"`
			Name      string     `json:"name"`
			Postcode  string     `json:"postcode"`
			District  string     `json:"district"`
			Length    int32      `json:"length,omitempty"`
			Lat       float64    `json:"lat"`
			Lon       float64    `json:"lon"`
			Relevance uint64     `json:"relevance"`
		}{
			ID:        p.ID,
			Class:     p.Class,
			Name:      p.Name,
			Postcode:  p.District.Postcode,
			District:  p.District.District,
			Length:    p.Length,
			Lat:       p.Lat,
			Lon:       p.Lon,
			Relevance: p.Relevance,
		})
	}
	if p.Class == locationClass {
		return json.Marshal(&struct {
			ID          int        `json:"id"`
			Class       PlaceClass `json:"class"`
			Type        string     `json:"type"`
			Name        string     `json:"name"`
			Street      string     `json:"street"`
			StreetID    int        `json:"streetID"`
			HouseNumber string     `json:"houseNumber"`
			Postcode    string     `json:"postcode"`
			District    string     `json:"district"`
			Lat         float64    `json:"lat"`
			Lon         float64    `json:"lon"`
			Relevance   uint64     `json:"relevance"`
		}{
			ID:          p.ID,
			Class:       p.Class,
			Type:        p.Type,
			Name:        p.Name,
			Street:      p.Street.Name,
			StreetID:    p.Street.ID,
			HouseNumber: p.HouseNumber,
			Postcode:    p.District.Postcode,
			District:    p.District.District,
			Lat:         p.Lat,
			Lon:         p.Lon,
			Relevance:   p.Relevance,
		})
	}
	if p.Class == houseNumberClass {
		return json.Marshal(&struct {
			ID          int        `json:"id"`
			Class       PlaceClass `json:"class"`
			Street      string     `json:"street"`
			StreetID    int        `json:"streetID"`
			HouseNumber string     `json:"houseNumber"`
			Postcode    string     `json:"postcode"`
			District    string     `json:"district"`
			Lat         float64    `json:"lat"`
			Lon         float64    `json:"lon"`
			Relevance   uint64     `json:"relevance"`
		}{
			ID:          p.ID,
			Class:       p.Class,
			Street:      p.Street.Name,
			StreetID:    p.Street.ID,
			HouseNumber: p.HouseNumber,
			Postcode:    p.District.Postcode,
			District:    p.District.District,
			Lat:         p.Lat,
			Lon:         p.Lon,
			Relevance:   p.Relevance,
		})
	}
	return []byte{}, fmt.Errorf("unexpected class '%s'", p.Type)
}

type result struct {
	Distance int    `json:"distance"`
	Place    *place `json:"place"`
}

// prefix represents precomputed completions and places for a given prefix
type prefix struct {

	// completions (i.e. places to suggest for this prefix (only if < maxPrefixLength)
	completions []*result

	// places covered by this prefix (if < maxPrefixLength those are the places in the completions)
	places []*place
}

type Places struct {

	// maximum prefix length
	maxPrefixLength int

	// the minimum number of completions to compute
	minCompletionCount int

	// the minimum input length before doing Levenshtein
	levMinimum int

	// distanceCut is used in result ranking. distanceCut is the delta in distances
	// to ignore in favor of relevance (unless one of the results has a distance of
	// 0).
	distanceCut int

	// duration to wait before evicting cache entries
	cacheTTLSeconds int
	cacheTTL        time.Duration

	// all places
	placesMap map[int]*place

	// a list of streets and locations (needed for completion)
	streetsAndLocations []*place

	// a map associating places with prefixes.
	prefixMap map[string]*prefix

	// cache for longer prefixes and prefixes with typo
	cache *ristretto.Cache

	// counts
	streetCount      int
	locationCount    int
	houseNumberCount int
	prefixCount      int

	// average lookup time
	m            sync.RWMutex
	avgQueryTime time.Duration
	queryCount   int64
}

type Metrics struct {
	MaxPrefixLength    int                `json:"maxPrefixLength"`
	MinCompletionCount int                `json:"minCompletionCount"`
	LevMinimum         int                `json:"levMinimum"`
	DistanceCut        int                `json:"distanceCut"`
	CacheTTL           int                `json:"cacheTTL"`
	StreetCount        int                `json:"streetCount"`
	LocationCount      int                `json:"locationCount"`
	HouseNumberCount   int                `json:"houseNumberCount"`
	PrefixCount        int                `json:"prefixCount"`
	CacheMetrics       *ristretto.Metrics `json:"cacheMetrics"`
	QueryCount         int64              `json:"queryCount"`
	AvgLookupTime      time.Duration      `json:"avgLookupTime"`
}

func NewPlaces(csvDistricts, csvStreets, csvLocations, csvHouseNumbers io.Reader) (*Places, error) {

	// basic init
	cacheTTLSeconds := viper.GetInt("CACHE_TTL")
	places := Places{
		maxPrefixLength:    viper.GetInt("MAX_PREFIX_LENGTH"),
		minCompletionCount: viper.GetInt("MIN_COMPLETION_COUNT"),
		levMinimum:         viper.GetInt("LEV_MINIMUM"),
		distanceCut:        viper.GetInt("RANKING_DISTANCE_CUT"),
		cacheTTLSeconds:    cacheTTLSeconds,
		cacheTTL:           time.Duration(cacheTTLSeconds) * time.Second,
	}

	// unmarshal district list
	var districtList []*district
	err := gocsv.Unmarshal(csvDistricts, &districtList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall districtList CSV data: %w", err)
	}

	// convert district list to map
	districtMap := make(map[string]*district)
	for _, d := range districtList {
		districtMap[d.Postcode] = d
	}

	// unmarshal street list
	var streets []*street
	err = gocsv.Unmarshal(csvStreets, &streets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall streets CSV data: %w", err)
	}

	// convert street list to place map (reassigning IDs and linking districts)
	places.placesMap = make(map[int]*place)
	var placeID int
	var streetID2placeID = make(map[int]int)
	for _, s := range streets {
		places.placesMap[placeID] = &place{
			ID:         placeID,
			Class:      streetClass,
			Name:       s.Name,
			cluster:    s.Cluster,
			District:   districtMap[s.Postcode],
			Lat:        s.Lat,
			Lon:        s.Lon,
			Length:     s.Length,
			simpleName: sanitizeString(s.Name),
		}
		streetID2placeID[s.ID] = placeID
		placeID += 1
		places.streetCount += 1
	}

	// unmarshal location list
	var locationList []*location
	err = gocsv.Unmarshal(csvLocations, &locationList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall locations CSV data: %w", err)
	}

	// extend places map by locations (assigning IDs, linking street-places and districts)
	for _, l := range locationList {
		streetPlace := places.placesMap[streetID2placeID[l.StreetID]]
		p := place{
			ID:          placeID,
			Class:       locationClass,
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
		places.locationCount += 1
	}

	var houseNumberList []*houseNumber
	err = gocsv.Unmarshal(csvHouseNumbers, &houseNumberList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall house numbers CSV data: %w", err)
	}

	// extend places map by house numbers (assigning IDs, linking street-places and districts)
	for _, h := range houseNumberList {
		streetPlace := places.placesMap[streetID2placeID[h.StreetID]]
		p := place{
			ID:          placeID,
			Class:       houseNumberClass,
			Street:      streetPlace,
			HouseNumber: h.HouseNumber,
			District:    districtMap[h.Postcode],
			Lat:         h.Lat,
			Lon:         h.Lon,
		}
		places.placesMap[placeID] = &p
		streetPlace.houseNumbers = append(streetPlace.houseNumbers, &p)
		placeID += 1
		places.houseNumberCount += 1
	}

	// collect streets and locations
	sl := make([]*place, places.streetCount+places.locationCount)
	i := 0
	for _, p := range places.placesMap {
		if p.Class != houseNumberClass {
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
	places.prefixCount = len(places.prefixMap)

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

func (bp *Places) GetCompletions(ctx context.Context, input string) []*result {
	start := time.Now()
	r := bp.getCompletions(ctx, input)
	go bp.updateMetrics(time.Since(start))
	return r
}

func (bp *Places) getCompletions(_ context.Context, input string) []*result {

	// dissect the input
	simpleInput := sanitizeString(input)
	runes := []rune(simpleInput)
	inputLength := len(runes)

	// if we have a matching cache entry return it
	cacheResults, hit := bp.cache.Get(simpleInput)
	if hit {
		if results, ok := cacheResults.([]*result); ok {

			// update relevance
			go bp.updateRelevance(results, simpleInput)

			return results
		} else {
			panic("failed to cast cache results")
		}
	}

	// if simpleInput is longer or equal to than maxPrefixLength
	if inputLength >= bp.maxPrefixLength {

		// compute the (max) prefix string
		prefixString := string(runes[:min(len(runes), bp.maxPrefixLength)])

		// if we have a matching entry in the prefix map
		if pf, ok := bp.prefixMap[prefixString]; ok {

			// do Levenshtein on the places associated with this prefix
			results := bp.levenshtein(pf.places, simpleInput)

			go func() {

				// update relevance
				bp.updateRelevance(results, simpleInput)

				// try to cache results (i.e. we extend the prefix map by longer prefixes)
				bp.cache.SetWithTTL(simpleInput, results, 0, bp.cacheTTL)
			}()

			return results

		} else {

			// do Levenshtein on all streets and locations
			results := bp.levenshtein(bp.streetsAndLocations, simpleInput)

			go func() {

				// update relevance
				bp.updateRelevance(results, simpleInput)

				// try to cache results (i.e. we extend the prefix map by long "faulty" prefixes)
				bp.cache.SetWithTTL(simpleInput, results, 0, bp.cacheTTL)
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

	// there is no matching prefix, but above levMinimum
	if inputLength >= bp.levMinimum {

		// do levenshtein on all streets and location
		results := bp.levenshtein(bp.streetsAndLocations, simpleInput)

		go func() {

			// update relevance for exact matches
			bp.updateRelevance(results, simpleInput)

			// try to cache results
			bp.cache.SetWithTTL(simpleInput, results, 0, bp.cacheTTL)
		}()

		return results
	}

	// as a last resort return the empty list
	return []*result{}
}

func (bp *Places) GetPlace(ctx context.Context, placeID int, houseNumber string) *place {
	start := time.Now()
	p := bp.getPlace(ctx, placeID, houseNumber)
	go bp.updateMetrics(time.Since(start))
	return p
}

func (bp *Places) getPlace(_ context.Context, placeID int, houseNumber string) *place {
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

func (bp *Places) Metrics() Metrics {
	bp.m.RLock()
	defer bp.m.RUnlock()
	return Metrics{
		MaxPrefixLength:    bp.maxPrefixLength,
		MinCompletionCount: bp.minCompletionCount,
		LevMinimum:         bp.levMinimum,
		DistanceCut:        bp.distanceCut,
		CacheTTL:           bp.cacheTTLSeconds,
		StreetCount:        bp.streetCount,
		LocationCount:      bp.locationCount,
		HouseNumberCount:   bp.houseNumberCount,
		PrefixCount:        bp.prefixCount,
		CacheMetrics:       bp.cache.Metrics,
		AvgLookupTime:      bp.avgQueryTime,
		QueryCount:         bp.queryCount,
	}
}

// computePrefixMap associates prefixes with completions xor places.
func (bp *Places) computePrefixMap() {

	pm := make(map[string]*prefix)

	for d := 1; d <= bp.maxPrefixLength; d++ {
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

			// append this place as a completion and place if below maxPrefixLength and
			// - its id exactly matches the current prefix or
			// - we don't have enough completions yet
			if d < bp.maxPrefixLength {
				if remainderLength == 0 || len(pm[prefixStr].completions) < bp.minCompletionCount {
					r := result{
						Distance: remainderLength,
						Place:    p,
					}
					pm[prefixStr].places = append(pm[prefixStr].places, p)
					pm[prefixStr].completions = append(pm[prefixStr].completions, &r)
					continue
				}
			} else {

				// we are at or above maxPrefixLength thus at the place as place
				pm[prefixStr].places = append(pm[prefixStr].places, p)
			}
		}

		bp.prefixMap = pm
	}
}

// updateRelevance increases the relevance for each exact match in the results
// slice and returns a slice containing the updated elements (if any).
func (bp *Places) updateRelevance(results []*result, simpleInput string) []*place {

	var updatedPlaces []*place

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
func (bp *Places) updateCompletions(updatedPlaces []*place) {

	simpleName := updatedPlaces[0].simpleName
	runes := []rune(simpleName)
	runesLen := len(runes)

	// assertion about updated place names
	for _, p := range updatedPlaces {
		if p.simpleName != simpleName {
			panic("unexpected place name")
		}
	}

	for d := 1; d < min(bp.maxPrefixLength, runesLen); d++ {

		prefixStr := string(runes[:d])

		// get the current completions for this prefix
		currentPlaces := bp.prefixMap[prefixStr].places

		// merge the results that where updated to the current completions and deduplicate
		mergedPlaces := deDuplicate(append(currentPlaces, updatedPlaces...))

		// do Levenshtein on the merged places wrt. the prefix string
		results := bp.levenshtein(mergedPlaces, prefixStr)

		var newCompletions []*result
		var newPlaces []*place
		for _, r := range results {
			if r.Place.simpleName == prefixStr || len(newCompletions) < bp.minCompletionCount {
				newCompletions = append(newCompletions, r)
				newPlaces = append(newPlaces, r.Place)
			}
		}

	}
}

func (bp *Places) levenshtein(places []*place, simpleInput string) []*result {

	// for each place compute the Levenshtein-Distance between its simple name and the given simple input
	results := make([]*result, len(places))
	for i, p := range places {
		results[i] = &result{
			Distance: levenshtein.ComputeDistance(simpleInput, p.simpleName),
			Place:    p,
		}
	}

	// sort results via place ranking.
	sort.Slice(results, func(i, j int) bool {
		return bp.resultRanking(results[i], results[j])
	})

	// compute the number of completions to return (i.e. all exact matches filled up to minCompletionCount)
	count := min(bp.minCompletionCount, len(results))
	for i := count; i < len(results); i++ {
		if results[i].Place.simpleName == simpleInput {

			// we are past minCompletionCount but still have an exact match, therefore add it
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
	bp.queryCount += 1
	if bp.avgQueryTime == 0 {
		bp.avgQueryTime = duration
	} else {
		bp.avgQueryTime = (bp.avgQueryTime + duration) / 2
	}
}

// resultRanking compares two levenshtein results wrt. distance, relevance, class, and (in case of streets) length.
// resultRanking returns true, if the first place should be ranked higher than the second one. resultRanking should
// be used in sorting slices (analog to the lesser function) sorting higher ranks to the beginning.
func (bp *Places) resultRanking(i, j *result) bool {

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

	// If none of the results is an exact match but the delta in distances is greater than distanceCut
	// the one with the smaller distance will be ranked higher.
	distanceDelta := abs(di - dj)
	if distanceDelta > bp.distanceCut {
		if di < dj {
			return true
		} else {
			return false
		}
	}

	pi := i.Place
	pj := j.Place

	// As there is no exact match and the delta in distances is within distanceCut,
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
		if pi.Class == streetClass {
			return true
		} else {
			return false
		}
	}

	// if streets, less by longer length (by the above clause, types must be equal)
	if pi.Class == streetClass {
		if pi.Length > pj.Length {
			return true
		} else {
			return false
		}
	}

	return false
}

// placeLesser compares two places wrt. string length and lexical order. placeLesser return true, if the first place
// is less than the second one.
func placeLesser(i, j *place) bool {

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
func deDuplicate(results []*place) []*place {

	ids := make(map[int]interface{})
	var places []*place

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
