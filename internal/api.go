package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"net/http"
	"strconv"
	"time"
)

type BerlinPlacesAPI struct {
	berlinPlaces *places.Places
}

// NewBerlinPlacesAPI initializes the BerlinPlacesAPI.
func NewBerlinPlacesAPI(berlinPlaces *places.Places) BerlinPlacesAPI {
	return BerlinPlacesAPI{
		berlinPlaces: berlinPlaces,
	}
}

// RegisterRoutes registers BerlinPlacesAPI routes.
func (bpa BerlinPlacesAPI) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/complete/", bpa.getCompletions)
	router.GET("/complete", bpa.getCompletions)
	router.GET("/place/:placeID/", bpa.getPlace)
	router.GET("/place/:placeID", bpa.getPlace)
}

func (bpa BerlinPlacesAPI) getCompletions(c *gin.Context) {

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
	results := bpa.berlinPlaces.GetCompletions(ctx, text)

	// return (i.e. forward) the response
	c.JSON(http.StatusOK, results)
}

func (bpa BerlinPlacesAPI) getPlace(c *gin.Context) {

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
	p := bpa.berlinPlaces.GetPlace(ctx, placeID, houseNumber)
	if p == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// return (i.e. forward) the response
	c.JSON(http.StatusOK, p)
}
