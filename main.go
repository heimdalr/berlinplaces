package main

import (
	"context"
	"fmt"
	"github.com/heimdalr/berlinplaces/internal"
	"github.com/heimdalr/berlinplaces/pkg/data"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

const viperEnvPrefix = "PLACES"

var buildVersion = "to be set by linker"
var buildGitHash = "to be set by linker"

// The application type.
type application struct {
	http.Server
}

// main.
func main() {

	// get Environment variables
	viperSetup()

	// configure the logger
	if viper.GetBool("DEBUG") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// show version info
	log.Info().Str("version", buildVersion).Str("hash", buildGitHash).Msg("app")

	// show config
	log.Info().
		Bool("debug", viper.GetBool("DEBUG")).
		Str("port", viper.GetString("PORT")).
		Bool("spec", viper.GetBool("SPEC")).
		Bool("demo", viper.GetBool("DEMO")).
		Msg("config")

	// initialize the app
	var app application
	err := app.initialize()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize app")
	}

	// run the app
	app.run()

	// wait for an interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// shutdown the app
	app.shutdown()
}

// Viper setup.
func viperSetup() {

	// bind env variables to viper keys
	viper.SetEnvPrefix(viperEnvPrefix)
	viper.AutomaticEnv()

	// default values
	viper.SetDefault("DEBUG", true)
	viper.SetDefault("PORT", "8080")

	// for places config set env defaults based on pkg defaults
	c := places.DefaultConfig
	viper.SetDefault("MAX_PREFIX_LENGTH", c.MaxPrefixLength)
	viper.SetDefault("MIN_COMPLETION_COUNT", c.MinCompletionCount)
	viper.SetDefault("MIN_LEV", c.MinLev)
	viper.SetDefault("DISTANCE_CUT", c.DistanceCut)
	viper.SetDefault("CACHE_TTL", c.CacheTTL)

	viper.SetDefault("DISTRICTS_CSV", "_data/districts.csv") // relative to project root
	viper.SetDefault("PLACES_CSV", "_data/places.csv")

	// set defaults for whether to enable swagger-docs depending on DEBUG
	if viper.GetBool("DEBUG") {
		viper.SetDefault("SPEC", true)
	} else {
		viper.SetDefault("SPEC", false)
	}

	// set defaults for whether to enable demo website depending on DEBUG
	if viper.GetBool("DEBUG") {
		viper.SetDefault("DEMO", true)
	} else {
		viper.SetDefault("DEMO", false)
	}

}

// initialize the application.
func (app *application) initialize() error {

	// initialize a router
	router := httprouter.New()

	// our panic Handler
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, params interface{}) {
		log.Error().Msgf("Caught panic: %v", params)
		log.Debug().Msgf("Stacktrace: %s", debug.Stack())
		w.WriteHeader(http.StatusInternalServerError)
	}

	// register swagger routes
	if viper.GetBool("SPEC") {
		router.ServeFiles("/swagger/*filepath", http.Dir("swagger"))
	}

	// register demo routes and redirect (if desired)
	if viper.GetBool("DEMO") {
		router.ServeFiles("/demo/*filepath", http.Dir("demo"))
		router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			http.Redirect(w, r, "/demo/", http.StatusTemporaryRedirect)
		})
	}

	// open (close) districts CSV file
	districtsFileName := viper.GetString("DISTRICTS_CSV")
	districtsReader, err := os.Open(districtsFileName)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", districtsFileName, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(districtsReader)

	// open (close) places CSV file
	placesFileName := viper.GetString("PLACES_CSV")
	placesReader, err := os.Open(placesFileName)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", placesFileName, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(placesReader)

	// init the data provider
	dataProvider := data.CSVProvider{
		DistrictsReader: districtsReader,
		PlacesReader:    placesReader,
	}

	// places configuration
	placesConfig := places.Config{
		MaxPrefixLength:    viper.GetInt("MAX_PREFIX_LENGTH"),
		MinCompletionCount: viper.GetInt("MIN_COMPLETION_COUNT"),
		MinLev:             viper.GetInt("MIN_LEV"),
		DistanceCut:        viper.GetInt("DISTANCE_CUT"),
		CacheTTL:           viper.GetDuration("CACHE_TTL"),
	}

	// initialize (berlin) places

	p, err := placesConfig.NewPlaces(dataProvider)
	if err != nil {
		panic(fmt.Errorf("failed to initialize places: %w", err))
	}

	// log basic stats about places
	metrics := p.Metrics()
	log.Info().
		Int32("streetCount", metrics.StreetCount).
		Int32("locationCount", metrics.LocationCount).
		Int32("houseNumberCount", metrics.HouseNumberCount).
		Int("prefixCount", metrics.PrefixCount).
		Msg("places")

	// register places routes
	placesAPI := internal.PlacesAPI{Places: p}
	router.GET("/places", placesAPI.GetCompletions)
	router.GET("/places/:placeID", placesAPI.GetPlace)
	router.GET("/metrics", placesAPI.GetMetrics)

	// version
	versionAPI := internal.VersionAPI{Version: buildVersion, Hash: buildGitHash}
	router.GET("/version", versionAPI.GetVersion)

	// wrap the router into a logging middleware
	loggingRouter := internal.LoggerMiddleware{Handler: router, Logger: log.Logger}

	// setup HTTP server
	app.Server = http.Server{
		Addr:           fmt.Sprintf(":%s", viper.GetString("PORT")),
		Handler:        loggingRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return nil
}

// run the application.
func (app *application) run() {

	// start the server (in a goroutine)
	go func() {
		if err := app.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server failed")
		}
	}()

	log.Info().Msgf("listening on http://localhost:%s", strings.TrimLeft(app.Server.Addr, ":"))
}

// shutdown shuts the application down.
func (app *application) shutdown() {

	// Gracefully shutdown server (in a timeout context).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	log.Info().Msg("gracefully shutting down")
	defer cancel()
	if err := app.Server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
	}
}
