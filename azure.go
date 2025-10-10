package godotenvsecrets

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

// Azure secrets decode function
func azureDecodeSecret(secretName string) (any, error) {

	vaultURI := os.Getenv("KEY_VAULT_URI")
	if vaultURI == "" {
		log.Fatal("KEY_VAULT_URI environment variable not set")
	}

	// 2. Authenticate using DefaultAzureCredential
	// This will try to authenticate using: Environment variables, Managed Identity, Azure CLI, etc.
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}

	// 3. Create the Secret Client
	client, err := azsecrets.NewClient(vaultURI, credential, nil)
	if err != nil {
		log.Fatalf("Failed to create secret client: %v", err)
	}

	// 4. Retrieve the secret
	// The context is used to manage the request's lifecycle (e.g., timeouts)
	ctx := context.Background()

	// Pass an empty string for the secret version to get the latest version
	version := ""
	resp, err := client.GetSecret(ctx, secretName, version, nil)
	if err != nil {
		log.Fatalf("Failed to get the secret '%s': %v", secretName, err)
	}

	// 5. Access and use the secret value
	secretValue := *resp.Value // The secret value is a pointer to a string
	fmt.Printf("Successfully retrieved secret '%s'.\n", secretName)
	fmt.Printf("Secret value: %s\n", secretValue)

	return secretValue, nil
}
