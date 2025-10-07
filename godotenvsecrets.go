package godotenvsecrets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/joho/godotenv"
)

type kv_t map[string]string
type secrets_t map[string]interface{}

var reEnvSecret *regexp.Regexp
var awsCache kv_t = make(kv_t)

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

		case "env":
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
			// 	// Implement Azure Key Vault retrieval here
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

func awsDecodeSecret(secretName string, secretKey string) (any, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION environment variable is not set")
	}

	// Check cache first
	if cached, ok := awsCache[secretName]; ok {
		var secrets secrets_t
		err := json.Unmarshal([]byte(cached), &secrets)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal cached secret: %v", err)
		}
		if val, ok := secrets[secretKey]; ok {
			fmt.Printf("Found in cache: @aws:%s:%s\n", secretName, secretKey)
			return val, nil
		}
		return "", fmt.Errorf("secret key '%s' not found in cached secret '%s'", secretKey, secretName)
	}

	fmt.Printf("Getting secret from aws: @aws:%s:%s\n", secretName, secretKey)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	secretString := *result.SecretString

	// Cache the raw secret string for future lookups
	awsCache[secretName] = secretString

	secrets := make(secrets_t)
	err = json.Unmarshal([]byte(secretString), &secrets)
	if err != nil {
		log.Fatal(err)
	}

	if val, ok := secrets[secretKey]; ok {
		return val, nil
	}
	return "", fmt.Errorf("secret key '%s' not found in secret '%s'", secretKey, secretName)
}
