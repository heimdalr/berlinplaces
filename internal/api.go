package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"net/http"
	"strconv"
	"time"
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
func (api PlacesAPI) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/complete/", api.getCompletions)
	router.GET("/complete", api.getCompletions)
	router.GET("/place/:placeID/", api.getPlace)
	router.GET("/place/:placeID", api.getPlace)
}

func (api PlacesAPI) getCompletions(c *gin.Context) {

	// timeout in seconds for calling geoapify
	const timeout = 5

	// get the search text from the request
	text := c.Query("text")
	if text == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// derive a timeout context
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	// query the geocoder
	results := api.places.GetCompletions(ctx, text)

	// return (i.e. forward) the response
	c.JSON(http.StatusOK, results)
}

func (api PlacesAPI) getPlace(c *gin.Context) {

	// timeout in seconds for calling geoapify
	const timeout = 5

	// parse the place ID
	placeIDStr := c.Param("placeID")
	var placeID int
	if placeIDStr != "" {
		id, err := strconv.Atoi(placeIDStr)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		placeID = id
	} else {
		c.Status(http.StatusBadRequest)
		return
	}
	// https://www.google.com/maps/place/52%C2%B045'92.1%22N+13%C2%B019'21.1%22E/@52.4592118,13.3222465,17.75z/
	// get the search houseNumber from the request (if any)
	houseNumber := c.Query("houseNumber")

	// derive a timeout context
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	// query the geocoder
	p := api.places.GetPlace(ctx, placeID, houseNumber)
	if p == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// return (i.e. forward) the response
	c.JSON(http.StatusOK, p)
}
