package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/heimdalr/berlinplaces/internal"
	"github.com/heimdalr/berlinplaces/pkg/places"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
	"net"
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

// The App type.
type App struct {
	Server http.Server
}

// main function to boot up everything.
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
	viper.SetDefault("DEBUG", true)
	viper.SetDefault("PORT", "8080")

	// length of prefixes to precompute
	// be careful this is close to exponential (2 => ~500, 3 => ~3000, 4 => ~10000 prefixes)
	viper.SetDefault("MAX_PREFIX_LENGTH", 4)
	viper.SetDefault("MIN_COMPLETION_COUNT", 10)
	viper.SetDefault("LEV_MINIMUM", 4)

	viper.SetDefault("DISTRICTS_CSV", "_data/districts.csv") // relative to project root
	viper.SetDefault("STREETS_CSV", "_data/streets.csv")
	viper.SetDefault("LOCATIONS_CSV", "_data/locations.csv")
	viper.SetDefault("HOUSE_NUMBERS_CSV", "_data/housenumbers.csv")

	viper.SetDefault("RANKING_DISTANCE_CUT", 4) // distanceCut is the delta in distances to ignore in favor of relevance

	viper.SetDefault("CACHE_TTL", 300) // number of seconds before evicting cache entries

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

// Initialize the application.
func (app *App) Initialize() error {

	// initialize a router
	router := httprouter.New()

	// our panic handler
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
			http.Redirect(w, r, "/demo/", 301)
		})
	}

	// initialize places
	p, err := initPlaces()
	if err != nil {
		return err
	}
	m := p.Metrics()
	log.Info().
		Int("streetCount", m.StreetCount).
		Int("locationCount", m.LocationCount).
		Int("houseNumberCount", m.HouseNumberCount).
		Int("prefixCount", m.PrefixCount).
		Msg("places")

	// register places routes
	internal.NewPlacesAPI(p).RegisterRoutes(router)

	// version
	router.GET("/version", getVersion)

	// wrap the router into a logging middleware
	lmw := loggerMiddleware{router}

	// setup HTTP server
	app.Server = http.Server{
		Addr:           fmt.Sprintf(":%s", viper.GetString("PORT")),
		Handler:        lmw,
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
	p, err := places.NewPlaces(districtsFile, streetsFile, locationsFile, houseNumbersFile)
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
		panic(fmt.Errorf("failed to marshall version info: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		panic(fmt.Errorf("failed to write response body: %w", err))
	}
}

// loggerMiddleware a type to implement our logging middleware (around the router).
type loggerMiddleware struct {
	handler http.Handler
}

// ServeHTTP implements the Handler interface for our logging middleware.
func (lmw loggerMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	url := *req.URL
	rw := negroni.NewResponseWriter(w)
	lmw.handler.ServeHTTP(rw, req)
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	log.Info().
		Str("remote", ip).
		Str("method", req.Method).
		Str("uri", url.RequestURI()).
		Int64("µs", time.Since(t).Microseconds()).
		Int("status", rw.Status()).
		Int("size", rw.Size()).
		Msg("request")
}
