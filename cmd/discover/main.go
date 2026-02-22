package main

import (
	"context"
	"fmt"
	"log"

	"snoozeql/internal/provider"
	"snoozeql/internal/provider/aws"
)

func main() {
	log.Println("Starting database discovery...")

	reg := provider.NewRegistry()

	// Register AWS provider with engineering-test credentials
	// TODO: Load from environment or config
	awsProvider, err := aws.NewRDSProvider("us-east-1", "", []string{}, "REDACTED_AWS_ACCESS_KEY", "REDACTED_AWS_SECRET_KEY")
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
