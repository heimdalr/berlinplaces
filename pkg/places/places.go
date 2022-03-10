package places

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/gocarina/gocsv"
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
			ID:       p.ID,
			Class:    p.Class,
			Name:     p.Name,
			Postcode: p.District.Postcode,
			District: p.District.District,
			Length:   p.Length,
			Lat:      p.Lat,
			Lon:      p.Lon,
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

	// all places
	placesMap map[int]*place

	// a list of streets and locations (needed for completion)
	streetsAndLocations []*place

	// counts
	streetCount      int
	locationCount    int
	houseNumberCount int

	// a map associating places with prefixes.
	prefixMap map[string]*prefix

	// cache for longer prefixes and prefixes with typo
	cache *ristretto.Cache

	// average lookup time
	m            sync.RWMutex
	avgQueryTime time.Duration
	queryCount   int64
}

type Metrics struct {
	MaxPrefixLength    int
	MinCompletionCount int
	LevMinimum         int
	StreetCount        int
	LocationCount      int
	HouseNumberCount   int
	PrefixCount        int
	CacheMetrics       *ristretto.Metrics
	QueryCount         int64
	AvgLookupTime      time.Duration
}

func NewPlaces(csvDistricts, csvStreets, csvLocations, csvHouseNumbers io.Reader, maxPrefixLength, minCompletionCount, levMinimum int) (*Places, error) {

	// basic init
	places := Places{
		maxPrefixLength:    maxPrefixLength,
		minCompletionCount: minCompletionCount,
		levMinimum:         levMinimum,
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

// placeLesser compares two places and returns true if the first is less (i.e. to
// be higher ranked in a list) than the second one.
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

	// if relevance differs, less by greater relevance
	if i.Relevance != j.Relevance {
		if i.Relevance > j.Relevance {
			return true
		} else {
			return false
		}
	}

	// if types differ streets over locations
	if i.Class != j.Class {
		if i.Class == streetClass {
			return true
		} else {
			return false
		}
	}

	// if streets, less by longer length (by the above clause, types must be equal)
	if i.Class == streetClass {
		if i.Length > j.Length {
			return true
		} else {
			return false
		}

	}

	// defaults
	return false
}

// computePrefixMap associates places with prefixes.
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

			// append this place as an exact match, if its id exactly matches the current prefix
			if remainderLength == 0 {

				r := result{
					Distance: 0,
					Place:    p,
				}
				pm[prefixStr].exact = append(pm[prefixStr].exact, &r)
				continue
			}

			// if not at max maxPrefixLength
			if d < bp.maxPrefixLength {

				// if completions are not yet full
				if len(pm[prefixStr].completions) < bp.minCompletionCount {
					r := result{
						Distance: remainderLength,
						Place:    p,
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

func (bp *Places) Metrics() Metrics {
	bp.m.RLock()
	defer bp.m.RUnlock()
	avgLookupTime := bp.avgQueryTime
	return Metrics{
		MaxPrefixLength:    bp.maxPrefixLength,
		MinCompletionCount: bp.minCompletionCount,
		LevMinimum:         bp.minCompletionCount,
		StreetCount:        bp.streetCount,
		LocationCount:      bp.locationCount,
		HouseNumberCount:   bp.houseNumberCount,
		PrefixCount:        len(bp.prefixMap),
		CacheMetrics:       bp.cache.Metrics,
		AvgLookupTime:      avgLookupTime,
		QueryCount:         bp.queryCount,
	}
}

func (bp *Places) updateQueryStats(duration time.Duration) {
	bp.m.Lock()
	defer bp.m.Unlock()
	bp.queryCount += 1
	if bp.avgQueryTime == 0 {
		bp.avgQueryTime = duration
	} else {
		bp.avgQueryTime = (bp.avgQueryTime + duration) / 2
	}
}

func (bp *Places) GetCompletions(ctx context.Context, input string) []*result {
	start := time.Now()
	r := bp.getCompletions(ctx, input)
	go bp.updateQueryStats(time.Since(start))
	return r
}

func (bp *Places) getCompletions(_ context.Context, input string) []*result {

	// dissect the input
	input = sanitizeString(input)
	runes := []rune(input)
	inputLength := len(runes)

	// if we have a matching cache entry return it
	cacheResults, hit := bp.cache.Get(input)
	if hit {
		if results, ok := cacheResults.([]*result); ok {
			return results
		} else {
			panic("failed to cast cache results")
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

			// do Levenshtein on all streets and locations
			results := bp.levenshtein(bp.streetsAndLocations, input)

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

		// do levenshtein on all streets and location
		results := bp.levenshtein(bp.streetsAndLocations, input)

		// try to cache results
		go func() {
			_ = bp.cache.Set(input, results, 0)
		}()

		return results
	}

	// as a last resort return the empty list
	return []*result{}
}

func (bp *Places) GetPlace(ctx context.Context, placeID int, houseNumber string) *place {
	start := time.Now()
	p := bp.getPlace(ctx, placeID, houseNumber)
	go bp.updateQueryStats(time.Since(start))
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

func (bp *Places) levenshtein(places []*place, text string) []*result {

	// for each placeOld compute the Levenshtein-Distance between its simple name and the given text
	results := make([]*result, len(places))
	for i, p := range places {
		results[i] = &result{
			Distance: levenshtein.ComputeDistance(text, p.simpleName),
			Place:    p,
		}
	}

	// sort the completions slice by Levenshtein-Distance and then lesser function
	sort.Slice(results, func(i, j int) bool {
		di := results[i].Distance
		dj := results[j].Distance
		return di < dj || (di == dj && placeLesser(results[i].Place, results[j].Place))
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
