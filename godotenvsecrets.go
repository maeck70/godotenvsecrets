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

var reEnvSecret *regexp.Regexp

type secrets_t map[string]string

func init() {
	// Matches secrets like @aws:dev/goenvsecrets:serviceaccount, @aws:dev/goenvsecrets/serviceaccount, etc.
	// Groups: provider, path, secret
	reEnvSecret = regexp.MustCompile(`^@([a-zA-Z0-9]+):([a-zA-Z0-9\-/]+)(?:[/:]([a-zA-Z0-9\-]+))?$`)

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

func Getenv(envkey string) (string, error) {
	var secretName string
	var secretKey string

	if envkey[0] == '@' {
		matches := reEnvSecret.FindStringSubmatch(envkey)

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

func awsDecodeSecret(secretName string, secretKey string) (string, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION environment variable is not set")
	}

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	// Your code goes here.
	secrets := make(secrets_t)
	err = json.Unmarshal([]byte(secretString), &secrets)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := secrets[secretKey]; !ok {
		return "", fmt.Errorf("secret key '%s' not found in secret '%s'", secretKey, secretName)
	}

	return secrets[secretKey], nil
}
