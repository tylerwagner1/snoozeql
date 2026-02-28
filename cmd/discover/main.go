package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"snoozeql/internal/provider"
	"snoozeql/internal/provider/aws"
)

func main() {
	log.Println("Starting database discovery...")

	// Load credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	if accessKey == "" || secretKey == "" {
		log.Fatal("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are required")
	}

	reg := provider.NewRegistry()

	awsProvider, err := aws.NewRDSProvider(region, "", []string{}, accessKey, secretKey)
	if err != nil {
		log.Fatalf("Failed to create AWS provider: %v", err)
	}
	reg.Register("aws", awsProvider)

	ctx := context.Background()
	instances, err := reg.ListAllDatabases(ctx)
	if err != nil {
		log.Fatalf("Failed to list databases: %v", err)
	}

	log.Printf("Discovered %d database instances", len(instances))

	for _, inst := range instances {
		fmt.Printf("- %s (%s/%s) - %s - Region: %s\n",
			inst.Name, inst.Provider, inst.Engine, inst.Status, inst.Region)
	}
}
