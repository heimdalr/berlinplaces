package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type PlacesAPI struct {
	*places.Places
}

func (placesAPI PlacesAPI) GetCompletions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// get the search text from the request
	queryValues := r.URL.Query()
	text := queryValues.Get("text")
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get completions
	results := placesAPI.Places.GetCompletions(context.Background(), text)

	// encode completions
	j, err := json.Marshal(results)
	if err != nil {
		panic(fmt.Errorf("failed to marshall results: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		panic(fmt.Errorf("failed to write response body: %w", err))
	}
}

func (placesAPI PlacesAPI) GetPlace(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// parse the place ID
	placeIDStr := ps.ByName("placeID")
	var placeID int64
	if placeIDStr != "" {
		id, err := strconv.ParseInt(placeIDStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		placeID = id
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the search houseNumber from the request (if any)
	queryValues := r.URL.Query()
	houseNumber := queryValues.Get("houseNumber")

	// get the place
	p := placesAPI.Places.GetPlace(context.Background(), placeID, houseNumber)
	if p == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// encode place
	j, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Errorf("failed to marshall place: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		panic(fmt.Errorf("failed to write response body: %w", err))
	}
}

func (placesAPI PlacesAPI) GetMetrics(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	m := placesAPI.Places.Metrics()
	j, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("failed to marshall metrics: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		panic(fmt.Errorf("failed to write response body: %w", err))
	}
}
