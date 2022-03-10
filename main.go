package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

	// get Environment variables
	viperSetup()

	// configure the logger
	if viper.GetString("MODE") == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// show version info
	log.Info().Str("version", buildVersion).Str("hash", buildGitHash).Msg("app info")

	// initialize the app
	var app App
	err := app.Initialize()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize app")
	}

	// run the app
	app.Run()

	// wait for an interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// shutdown the app
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

	viper.SetDefault("MAX_PREFIX_LENGTH", 6)
	viper.SetDefault("MIN_COMPLETION_COUNT", 6)
	viper.SetDefault("LEV_MINIMUM", 0)

	viper.SetDefault("DISTRICTS_CSV", "_data/districts.csv") // relative to project root
	viper.SetDefault("STREETS_CSV", "_data/streets.csv")
	viper.SetDefault("LOCATIONS_CSV", "_data/locations.csv")
	viper.SetDefault("HOUSE_NUMBERS_CSV", "_data/housenumbers.csv")

}

// Initialize the application.
func (app *App) Initialize() error {

	// initialize a router.
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, params interface{}) {
		log.Error().Msgf("Caught panic: %v", params)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// register swagger routes
	router.ServeFiles("/swagger/*filepath", http.Dir("swagger"))

	// register web routes
	router.ServeFiles("/web/*filepath", http.Dir("web"))

	/*// initialize places
	p, err := initPlaces()
	if err != nil {
		return err
	}
	m := p.Metrics()
	log.Info().
		Int("maxPrefixLength", m.MaxPrefixLength).
		Int("minCompletionCount", m.MinCompletionCount).
		Int("levMinimum", m.LevMinimum).
		Int("streetCount", m.StreetCount).
		Int("locationCount", m.LocationCount).
		Int("houseNumberCount", m.HouseNumberCount).
		Int("prefixCount", m.PrefixCount).
		Msg("initialized places")

	// register places routes
	internal.NewPlacesAPI(p).RegisterRoutes(router)
	*/
	// version
	router.GET("/version", getVersion)

	// redirect / to /web
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/web/", 301)
	})

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

	// start the server (in a goroutine)
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

// initPlaces opens the necessary CSV files and initializes places.
func initPlaces() (*places.Places, error) {

	// open the districts CSV
	districtsFile, errDF := os.Open(viper.GetString("DISTRICTS_CSV"))
	if errDF != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), errDF)
	}
	defer func() {
		_ = districtsFile.Close()
	}()

	// open the streets CSV
	streetsFile, errSF := os.Open(viper.GetString("STREETS_CSV"))
	if errSF != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), errSF)
	}
	defer func() {
		_ = streetsFile.Close()
	}()

	// open the locations CSV
	locationsFile, errLF := os.Open(viper.GetString("LOCATIONS_CSV"))
	if errLF != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), errLF)
	}
	defer func() {
		_ = locationsFile.Close()
	}()

	// open the house numbers CSV
	houseNumbersFile, errHNF := os.Open(viper.GetString("HOUSE_NUMBERS_CSV"))
	if errHNF != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), errHNF)
	}
	defer func() {
		_ = houseNumbersFile.Close()
	}()

	// initialize places
	maxPrefixLength := viper.GetInt("MAX_PREFIX_LENGTH")
	minCompletionCount := viper.GetInt("MIN_COMPLETION_COUNT")
	levMinimum := viper.GetInt("LEV_MINIMUM")

	p, err := places.NewPlaces(districtsFile, streetsFile, locationsFile, houseNumbersFile, maxPrefixLength, minCompletionCount, levMinimum)
	if err != nil {
		panic(fmt.Errorf("failed to initialize places: %w", err))
	}

	return p, nil
}

// getVersion is the handler for the /version-endpoint.
func getVersion(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	versionInfo := struct {
		Version string `json:"version"`
		Hash    string `json:"hash"`
	}{buildVersion, buildGitHash}
	j, err := json.Marshal(versionInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(j)
}
