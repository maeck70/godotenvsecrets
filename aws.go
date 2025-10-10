package godotenvsecrets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var awsCache kv_t = make(kv_t)

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
