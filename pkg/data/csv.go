package data

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"os"
	"strings"
)

type CSVProvider struct {
	DistrictsFile    string
	StreetsFile      string
	LocationsFile    string
	HouseNumbersFile string
}

// normalizer is the function to apply to headers and struct fields before trying to match.
// see: https://pkg.go.dev/github.com/gocarina/gocsv#Normalizer
func normalizer(s string) string {
	return strings.Trim(strings.ToLower(s), "_")
}

func (p CSVProvider) Get() (*Data, error) {

	gocsv.SetHeaderNormalizer(normalizer)

	d := Data{}

	jobs := []struct {
		fileName string
		data     interface{}
	}{
		{p.DistrictsFile, &d.Districts},
		{p.StreetsFile, &d.Streets},
		{p.LocationsFile, &d.Locations},
		{p.HouseNumbersFile, &d.HouseNumbers},
	}

	for _, j := range jobs {

		err := unmarshall(j.fileName, j.data)
		if err != nil {
			return nil, err
		}
	}

	return &d, nil
}

func unmarshall(fileName string, data interface{}) error {

	// open the CSV file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", fileName, err)
	}

	// unmarshall into given interface
	err = gocsv.Unmarshal(file, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshall '%s' data: %w", fileName, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to clode '%s': %w", fileName, err)
	}

	return nil
}
