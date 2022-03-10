package internal

import (
	"context"
	"encoding/json"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type PlacesAPI struct {
	places *places.Places
}

// NewPlacesAPI initializes the PlacesAPI.
func NewPlacesAPI(p *places.Places) PlacesAPI {
	return PlacesAPI{
		places: p,
	}
}

// RegisterRoutes registers PlacesAPI routes.
func (api PlacesAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/api/complete/", api.getCompletions)
	router.GET("/api/complete", api.getCompletions)
	router.GET("/api/place/:placeID/", api.getPlace)
	router.GET("/api/place/:placeID", api.getPlace)
	router.GET("/api/metrics/", api.getMetrics)
	router.GET("/api/metrics", api.getMetrics)
}

func (api PlacesAPI) getCompletions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// get the search text from the request
	queryValues := r.URL.Query()
	text := queryValues.Get("text")
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get completions
	results := api.places.GetCompletions(context.Background(), text)

	// encode completions
	j, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(j)
}

func (api PlacesAPI) getPlace(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// parse the place ID
	placeIDStr := ps.ByName("placeID")
	var placeID int
	if placeIDStr != "" {
		id, err := strconv.Atoi(placeIDStr)
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
	p := api.places.GetPlace(context.Background(), placeID, houseNumber)
	if p == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// encode place
	j, err := json.Marshal(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(j)
}

func (api PlacesAPI) getMetrics(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	m := api.places.Metrics()
	j, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(j)
}
