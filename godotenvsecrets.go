package godotenvsecrets

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type kv_t map[string]string
type secrets_t map[string]interface{}

var reEnvSecret *regexp.Regexp

func init() {
	// Matches secrets like @aws:dev/goenvsecrets:serviceaccount, @aws:dev/goenvsecrets/serviceaccount, etc.
	// Groups: provider, path, secret
	// reEnvSecret = regexp.MustCompile(`^@([a-zA-Z0-9]+):([a-zA-Z0-9\-/]+)(?:[/:]([a-zA-Z0-9\-]+))?$`)
	reEnvSecret = regexp.MustCompile(`^@([a-zA-Z0-9_\-]+):([a-zA-Z0-9_\-/]+)(?:[/:]([a-zA-Z0-9_\-]+))?$`)

	/* Examples:
	@aws:dev/goenvsecrets:serviceaccount
	@aws:dev/goenvsecrets/serviceaccount
	@aws:dev/goenvsecrets/more:serviceaccount
	@azure:dev/goenvsecrets:serviceaccount
	@env:ENVVARIABLE = @env ENVVARIABLE
	*/
}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}

func Getenv(envkey string) (any, error) {
	var secretName string
	var secretKey string

	if envkey[0] == '@' {
		matches := reEnvSecret.FindStringSubmatch(envkey)

		if len(matches) == 0 {
			return "", fmt.Errorf("invalid secret format, invalid characters (allowed: a-z A-Z 0-9 _ -) or structure")
		}

		provider := strings.ToLower(matches[1])

		switch provider {
		case "aws":
			secretName = matches[2]
			secretKey = matches[3]

			if secretKey == "" {
				return "", fmt.Errorf("invalid secret format, missing secret key")
			}

		case "env", "azure":
			secretName = matches[2]

		default:
			return "", fmt.Errorf("secrets provider %s not implemented", provider)
		}

		if envkey[0] == '@' {

			switch provider {
			case "aws":
				// Format: @aws:secret-name:secret-key
				secretValue, err := awsDecodeSecret(secretName, secretKey)
				return secretValue, err
			case "env":
				// Format: @env:ENVVARIABLE
				return os.Getenv(secretName), nil
			// case "azure":
			// 	secretValue, err := azureDecodeSecret(secretName)
			// 	return secretValue, err
			// case: "gcp":
			// 	// Implement Google Secret Manager retrieval here
			// case "vault":
			// 	// Implement HashiCorp Vault retrieval here
			// case "k8s":
			// 	// Implement Kubernetes Secrets retrieval here
			// case "1password":
			// 	// Implement 1Password retrieval here
			default:
				return "", fmt.Errorf("secrets provider %s not implemented", provider)
			}
		}
	}

	secretValue := os.Getenv(envkey)
	if secretValue == "" {
		return "", fmt.Errorf("environment variable '%s' not set", envkey)
	}

	return secretValue, nil
}
