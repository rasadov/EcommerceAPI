package config

import (
	"os"
	"strconv"
	"strings"
)

var (
	AccountUrl          string
	ProductUrl          string
	OrderUrl            string
	PaymentUrl          string
	RecommenderUrl      string
	SecretKey           string
	Issuer              string
	Environment         string
	EnableIntrospection bool
)

// LoadConfig loads or reloads configuration from environment variables.
func LoadConfig() {
	AccountUrl = os.Getenv("ACCOUNT_SERVICE_URL")
	ProductUrl = os.Getenv("PRODUCT_SERVICE_URL")
	OrderUrl = os.Getenv("ORDER_SERVICE_URL")
	PaymentUrl = os.Getenv("PAYMENT_SERVICE_URL")
	RecommenderUrl = os.Getenv("RECOMMENDER_SERVICE_URL")
	SecretKey = os.Getenv("SECRET_KEY")
	Issuer = os.Getenv("ISSUER")

	// Determine environment: ENVIRONMENT > APP_ENV > ENV, defaulting to "development"
	Environment = os.Getenv("ENVIRONMENT")
	if Environment == "" {
		Environment = os.Getenv("APP_ENV")
	}
	if Environment == "" {
		Environment = os.Getenv("ENV")
	}
	if Environment == "" {
		Environment = "development"
	}

	EnableIntrospection = resolveIntrospection(Environment)
}

func isProductionEnv(env string) bool {
	normalized := strings.ToLower(strings.TrimSpace(env))
	return normalized == "production" || normalized == "prod"
}

// resolveIntrospection determines whether introspection should be enabled.
// If ENABLE_INTROSPECTION or INTROSPECTION_ENABLED is explicitly set, it takes precedence.
// Otherwise, introspection is disabled in production and enabled in other environments (e.g. development, local, test).
func resolveIntrospection(env string) bool {
	introspectionEnv := os.Getenv("ENABLE_INTROSPECTION")
	if introspectionEnv == "" {
		introspectionEnv = os.Getenv("INTROSPECTION_ENABLED")
	}

	if introspectionEnv != "" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(introspectionEnv))
		if err == nil {
			return parsed
		}
	}

	return !isProductionEnv(env)
}

// IsProduction returns true if the current environment is production.
func IsProduction() bool {
	return isProductionEnv(Environment)
}

func init() {
	LoadConfig()
}
