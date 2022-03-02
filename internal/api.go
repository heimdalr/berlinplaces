package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/heimdalr/berlinplaces/berlinplaces"
	"net/http"
	"time"
)

type BerlinPlacesAPI struct {
	berlinPlaces *berlinplaces.BerlinPlaces
}

// NewBerlinPlacesAPI initializes the BerlinPlacesAPI.
func NewBerlinPlacesAPI(berlinPlaces *berlinplaces.BerlinPlaces) BerlinPlacesAPI {
	return BerlinPlacesAPI{
		berlinPlaces: berlinPlaces,
	}
}

// RegisterRoutes registers BerlinPlacesAPI routes.
func (bpa BerlinPlacesAPI) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", bpa.get)
}

func (bpa BerlinPlacesAPI) get(c *gin.Context) {

	// timeout in seconds for calling geoapify
	const timeout = 5

	// get the search text from the syclist request
	text := c.Query("text")
	if text == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// derive a timeout context
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	// query the geocoder
	results := bpa.berlinPlaces.Query(ctx, text)

	// return (i.e. forward) the response
	c.JSON(http.StatusOK, results)
}
