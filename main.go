package main

import (
	"context"
	"fmt"
	"github.com/dn365/gin-zerolog"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heimdalr/berlinplaces/internal"
	"github.com/heimdalr/berlinplaces/pkg/places"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const viperEnvPrefix = "BP"

var buildVersion = "to be set by linker"
var buildGitHash = "to be set by linker"

// The App type.
type App struct {
	Server http.Server
}

// main function to boot up everything.
func main() {

	// Show version info.
	log.Info().Str("version", buildVersion).Str("hash", buildGitHash).Msg("")

	// Get Environment variables.
	viperSetup()

	// Set the gin mode.
	if viper.GetString("MODE") == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = ioutil.Discard
	}

	// ensure we have a valid config
	err := viperValidate()
	if err != nil {
		log.Fatal().Err(err).Msg("config validation failed")
	}

	var app App
	err = app.Initialize()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize app")
	}
	app.Run()

	// Wait for a interrupt.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	app.Shutdown()
}

// Viper setup.
func viperSetup() {

	// bind env variables to viper keys

	viper.SetEnvPrefix(viperEnvPrefix)
	viper.AutomaticEnv()

	// default values
	viper.SetDefault("MODE", "debug") // debug->debug or anything for release
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("CSV", "berlin.csv") // relative to project root

}

// viperValidate ensure a valid configuration
func viperValidate() error {
	if viper.GetString("CSV") == "" {
		return fmt.Errorf("missing configuration for ('%s_CSV')", viperEnvPrefix)
	}
	return nil
}

// Initialize the application.
func (app *App) Initialize() error {

	// initialize a router.
	router := gin.Default()

	// add a zerolog middleware.
	router.Use(ginzerolog.Logger("gin"))

	// register swagger routes
	router.StaticFS("swagger/", http.Dir("swagger"))

	// register web routes
	router.StaticFS("web/", http.Dir("web"))

	// open the berlin places csv
	file, err := os.Open(viper.GetString("CSV"))
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), err)
	}
	defer func() {
		_ = file.Close()
	}()

	// initialize places
	maxPrefixLength := 6
	minCompletionCount := 6
	levMinimum := 0
	berlinPlaces, err := places.NewPlaces(file, maxPrefixLength, minCompletionCount, levMinimum)
	if err != nil {
		panic(fmt.Errorf("failed to initialize berlinPlaces: %w", err))
	}
	log.Debug().
		Int("maxPrefixLength", maxPrefixLength).
		Int("minCompletionCount", minCompletionCount).
		Int("levMinimum", levMinimum).
		Int("placeCount", berlinPlaces.Metrics().PlaceCount).
		Int("prefixCount", berlinPlaces.Metrics().PrefixCount).
		Msg("initialized places")

	// register places routes
	internal.NewBerlinPlacesAPI(berlinPlaces).RegisterRoutes(router.Group("/api"))

	// version
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": buildVersion, "hash": buildGitHash})
	})

	// redirect / to /web
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web")
	})

	// use a CORS middleware (allow all).
	router.Use(cors.Default())

	// setup HTTP server
	app.Server = http.Server{
		Addr:           fmt.Sprintf(":%s", viper.GetString("PORT")),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return nil
}

// Run the application.
func (app *App) Run() {

	// Start the server in a goroutine (i.e. concurrently).
	go func() {
		if err := app.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server failed")
		}
	}()

	log.Info().Msgf("listening on http://localhost:%s", strings.TrimLeft(app.Server.Addr, ":"))
}

// Shutdown the application.
func (app *App) Shutdown() {

	// Gracefully shutdown server (in a timeout context).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	log.Info().Msg("gracefully shutting down")
	defer cancel()
	if err := app.Server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
	}
}
