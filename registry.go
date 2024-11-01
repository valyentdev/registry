package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	"github.com/distribution/distribution/v3/registry/auth"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	checkEnvironmentVariables()

	registry, err := setupRegistry()
	if err != nil {
		log.Fatal("some error occurred while setting up the registry", "err", err)
	}
	err = registry.ListenAndServe()
	if err != nil {
		log.Fatal("some error occurred while starting the registry", "err", err)
	}
}

func checkEnvironmentVariables() {
	requiredVars := []string{
		"REGISTRY_RAVEL_PASSWORD",
		"REGISTRY_HTTP_ADDR",
		"REGISTRY_HTTP_SECRET",
		"REGISTRY_STORAGE_S3_ACCESS_KEY",
		"REGISTRY_STORAGE_S3_SECRET_KEY",
		"REGISTRY_STORAGE_S3_BUCKET",
		"REGISTRY_STORAGE_S3_ENDPOINT",
		"VALYENT_API_BASE_URL",
	}
	for _, v := range requiredVars {
		checkEnvironmentVariablePresence(v)
	}
}

func checkEnvironmentVariablePresence(key string) {
	if os.Getenv(key) == "" {
		panic(fmt.Sprintf("%s env variable is not set", key))
	}
}

func setupRegistry() (*registry.Registry, error) {
	// Register the authentication scheme.
	if err := auth.Register("valyent", auth.InitFunc(newAccessController)); err != nil {
		return nil, err
	}

	config := &configuration.Configuration{}
	config.Log.Level = "info"
	config.HTTP.Addr = os.Getenv("REGISTRY_HTTP_ADDR")
	config.HTTP.Secret = os.Getenv("REGISTRY_HTTP_SECRET")
	config.HTTP.DrainTimeout = time.Duration(10) * time.Second
	config.HTTP.Net = "tcp"
	config.Auth = configuration.Auth{"valyent": configuration.Parameters{}}
	config.Storage = configuration.Storage{
		"s3": configuration.Parameters{
			"accesskey":      os.Getenv("REGISTRY_STORAGE_S3_ACCESS_KEY"),
			"secretkey":      os.Getenv("REGISTRY_STORAGE_S3_SECRET_KEY"),
			"bucket":         os.Getenv("REGISTRY_STORAGE_S3_BUCKET"),
			"endpoint":       os.Getenv("REGISTRY_STORAGE_S3_ENDPOINT"),
			"region":         os.Getenv("REGISTRY_STORAGE_S3_REGION"),
			"regionendpoint": os.Getenv("REGISTRY_STORAGE_S3_ENDPOINT"),
			"forcepathstyle": true,
		},
	}

	return registry.NewRegistry(context.Background(), config)
}
