package godotenvsecrets

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// Test case type for table-driven tests
type getenvTestCase struct {
	name        string
	loadErr     error
	getenvPath  string
	getenvValue string
	getenvErr   error
}

var testSet = []getenvTestCase{
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
		getenvValue: "",
		getenvErr:   fmt.Errorf("invalid secret format, missing secret key"),
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
		getenvErr:   fmt.Errorf("secret key '%s' not found in cached secret '%s'", "idonotexist", "dev/goenvsecrets"),
	},
	{
		name:        "Azure Not Implemented",
		loadErr:     nil,
		getenvPath:  "@azure:dev/goenvsecrets:notimplemented",
		getenvValue: "",
		getenvErr:   fmt.Errorf("secrets provider azure not implemented"),
	},
	{
		name:        "Env Var as a reference",
		loadErr:     nil,
		getenvPath:  "@env:ENVVARIABLE",
		getenvValue: "IAmNotASecret",
		getenvErr:   nil,
	},
	{
		name:        "Straight up Env Var",
		loadErr:     nil,
		getenvPath:  "ENVVARIABLE",
		getenvValue: "IAmNotASecret",
		getenvErr:   nil,
	},
	{
		name:        "Non set Env Var",
		loadErr:     nil,
		getenvPath:  "NOTINENV",
		getenvValue: "",
		getenvErr:   fmt.Errorf("environment variable 'NOTINENV' not set"),
	},
	{
		name:        "RabbitMQ Test",
		loadErr:     nil,
		getenvPath:  "@env:RABBITMQ_HOST",
		getenvValue: "localhost",
		getenvErr:   nil,
	},
	{
		name:        "RabbitMQ Test",
		loadErr:     nil,
		getenvPath:  "@env:RABBITMQ=HOST",
		getenvValue: "",
		getenvErr:   fmt.Errorf("invalid secret format, invalid characters (allowed: a-z A-Z 0-9 _ -) or structure"),
	},
	// {
	// 	name:        "Azure mysql password test",
	// 	loadErr:     nil,
	// 	getenvPath:  "@azure:mysql",
	// 	getenvValue: "hello",
	// 	getenvErr:   nil,
	// },
}

// Test each testSet case individually using the helper
func TestGetenv_SingleCases(t *testing.T) {
	os.Setenv("ENVVARIABLE", "IAmNotASecret")
	os.Setenv("RABBITMQ_HOST", "localhost")

	err := godotenv.Load()
	if err != nil {
		t.Fatalf("Failed to load .env: %v", err)
	}

	for _, tc := range testSet {
		t.Run(tc.name, func(t *testing.T) {
			runTestSetCase(t, tc)
		})
	}
}

// Helper to run a single test case from testSet
func runTestSetCase(t *testing.T, tc getenvTestCase) {
	var errStr, getEnvErrStr string

	got, err := Getenv(tc.getenvPath)

	if err != nil {
		errStr = err.Error()
	} else {
		errStr = ""
	}

	if tc.getenvErr != nil {
		getEnvErrStr = tc.getenvErr.Error()
	} else {
		getEnvErrStr = ""
	}

	if getEnvErrStr != errStr {
		t.Errorf("Expected error '%v', got '%v'", tc.getenvErr, err)
	}

	if got != tc.getenvValue {
		t.Errorf("Expected value '%s', got '%s'", tc.getenvValue, got)
	}
}
