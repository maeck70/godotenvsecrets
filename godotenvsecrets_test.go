package godotenvsecrets

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var testSet = []struct {
	name        string
	loadErr     error
	getenvPath  string
	getenvValue string
	getenvErr   error
}{
	{
		name:        "AWS Service Account",
		loadErr:     nil,
		getenvPath:  "@aws:dev/goenvsecrets:serviceaccount",
		getenvValue: "marcel",
		getenvErr:   nil,
	},
	{
		name:        "AWS Service Account",
		loadErr:     nil,
		getenvPath:  "@aws:dev/goenvsecrets/serviceaccount",
		getenvValue: "marcel",
		getenvErr:   nil,
	},
	{
		name:        "AWS Secret Key",
		loadErr:     nil,
		getenvPath:  "@aws:dev/goenvsecrets:secretkey",
		getenvValue: "supersecret",
		getenvErr:   nil,
	},
	{
		name:        "AWS Key Does Not Exist",
		loadErr:     nil,
		getenvPath:  "@aws:dev/goenvsecrets:idonotexist",
		getenvValue: "",
		getenvErr:   fmt.Errorf("secret key '%s' not found in secret '%s'", "idonotexist", "dev/goenvsecrets"),
	},
	{
		name:        "Azure Not Implemented",
		loadErr:     nil,
		getenvPath:  "@azure:dev/goenvsecrets:notimplemented",
		getenvValue: "",
		getenvErr:   fmt.Errorf("secrets provider not implemented"),
	},
	{
		name:        "Straight up Env Var",
		loadErr:     nil,
		getenvPath:  "ENVVARIABLE",
		getenvValue: "IAmNotASecret",
		getenvErr:   nil,
	},
}

// Table-driven test for Getenv using testSet
func TestGetenv_TableDriven(t *testing.T) {
	// Set up environment variable for "Straight up Env Var" test
	os.Setenv("ENVVARIABLE", "IAmNotASecret")

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, tc := range testSet {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Getenv(tc.getenvPath)
			if tc.getenvErr != nil {
				if err == nil || err.Error() != tc.getenvErr.Error() {
					t.Errorf("Expected error '%v', got '%v'", tc.getenvErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tc.getenvValue {
					t.Errorf("Expected value '%s', got '%s'", tc.getenvValue, got)
				}
			}
		})
	}
}
