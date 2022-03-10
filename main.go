package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/heimdalr/berlinplaces/internal"
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

	// Show version info.
	log.Info().Str("version", buildVersion).Str("hash", buildGitHash).Msg("")

	// Get Environment variables.
	viperSetup()

	// Set the gin mode.
	if viper.GetString("MODE") == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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

	viper.SetDefault("MAX_PREFIX_LENGTH", 6)
	viper.SetDefault("MIN_COMPLETION_COUNT", 6)
	viper.SetDefault("LEV_MINIMUM", 0)

	viper.SetDefault("DISTRICTS_CSV", "_data/districts.csv") // relative to project root
	viper.SetDefault("STREETS_CSV", "_data/streets.csv")
	viper.SetDefault("LOCATIONS_CSV", "_data/locations.csv")
	viper.SetDefault("HOUSENUMBERS_CSV", "_data/housenumbers.csv")

}

// viperValidate ensure a valid configuration
func viperValidate() error {
	return nil
}

// Initialize the application.
func (app *App) Initialize() error {

	// initialize a router.
	router := httprouter.New()

	// add a zerolog middleware.
	//router.Use(ginzerolog.Logger("gin"))

	// register swagger routes
	router.ServeFiles("/swagger/*filepath", http.Dir("swagger"))

	// register web routes
	router.ServeFiles("/web/*filepath", http.Dir("web"))

	// initialize places
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

	// version
	router.GET("/version", getVersion)

	// redirect / to /web
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/web/", 301)
	})

	// use a CORS middleware (allow all).
	//router.Use(cors.Default())

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

	// open the housnumbers CSV
	housnumbersFile, errHNF := os.Open(viper.GetString("HOUSENUMBERS_CSV"))
	if errHNF != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", viper.GetString("CSV"), errHNF)
	}
	defer func() {
		_ = housnumbersFile.Close()
	}()

	// initialize places
	maxPrefixLength := viper.GetInt("MAX_PREFIX_LENGTH")
	minCompletionCount := viper.GetInt("MIN_COMPLETION_COUNT")
	levMinimum := viper.GetInt("LEV_MINIMUM")

	p, err := places.NewPlaces(districtsFile, streetsFile, locationsFile, housnumbersFile, maxPrefixLength, minCompletionCount, levMinimum)
	if err != nil {
		panic(fmt.Errorf("failed to initialize places: %w", err))
	}

	return p, nil
}

// getVersion is the handler for /version
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
