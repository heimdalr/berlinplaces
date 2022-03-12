package internal

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type VersionAPI struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
}

// GetVersion is the Handler for the /version-endpoint.
func (versionAPI VersionAPI) GetVersion(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	j, err := json.Marshal(versionAPI)
	if err != nil {
		panic(fmt.Errorf("failed to marshall version info: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		panic(fmt.Errorf("failed to write response body: %w", err))
	}
}
