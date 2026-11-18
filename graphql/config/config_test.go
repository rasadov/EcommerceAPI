package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func clearConfigEnv() {
	os.Unsetenv("ACCOUNT_SERVICE_URL")
	os.Unsetenv("PRODUCT_SERVICE_URL")
	os.Unsetenv("ORDER_SERVICE_URL")
	os.Unsetenv("PAYMENT_SERVICE_URL")
	os.Unsetenv("RECOMMENDER_SERVICE_URL")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("ISSUER")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("ENV")
	os.Unsetenv("ENABLE_INTROSPECTION")
	os.Unsetenv("INTROSPECTION_ENABLED")
}

func TestConfigDefaults(t *testing.T) {
	clearConfigEnv()
	defer clearConfigEnv()

	LoadConfig()

	assert.Equal(t, "development", Environment)
	assert.True(t, EnableIntrospection, "introspection should be enabled by default in development")
	assert.False(t, IsProduction())
}

func TestConfigEnvironmentResolution(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedEnv    string
		expectedIntro  bool
		expectedIsProd bool
	}{
		{
			name:           "explicit development",
			envVars:        map[string]string{"ENVIRONMENT": "development"},
			expectedEnv:    "development",
			expectedIntro:  true,
			expectedIsProd: false,
		},
		{
			name:           "explicit dev",
			envVars:        map[string]string{"ENVIRONMENT": "dev"},
			expectedEnv:    "dev",
			expectedIntro:  true,
			expectedIsProd: false,
		},
		{
			name:           "explicit local",
			envVars:        map[string]string{"ENVIRONMENT": "local"},
			expectedEnv:    "local",
			expectedIntro:  true,
			expectedIsProd: false,
		},
		{
			name:           "explicit production",
			envVars:        map[string]string{"ENVIRONMENT": "production"},
			expectedEnv:    "production",
			expectedIntro:  false,
			expectedIsProd: true,
		},
		{
			name:           "explicit prod",
			envVars:        map[string]string{"ENVIRONMENT": "prod"},
			expectedEnv:    "prod",
			expectedIntro:  false,
			expectedIsProd: true,
		},
		{
			name:           "production case insensitive with spaces",
			envVars:        map[string]string{"ENVIRONMENT": "  Production  "},
			expectedEnv:    "  Production  ",
			expectedIntro:  false,
			expectedIsProd: true,
		},
		{
			name:           "fallback to APP_ENV",
			envVars:        map[string]string{"APP_ENV": "production"},
			expectedEnv:    "production",
			expectedIntro:  false,
			expectedIsProd: true,
		},
		{
			name:           "fallback to ENV",
			envVars:        map[string]string{"ENV": "production"},
			expectedEnv:    "production",
			expectedIntro:  false,
			expectedIsProd: true,
		},
		{
			name: "ENVIRONMENT takes precedence over APP_ENV and ENV",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"APP_ENV":     "production",
				"ENV":         "production",
			},
			expectedEnv:    "development",
			expectedIntro:  true,
			expectedIsProd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv()
			defer clearConfigEnv()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			LoadConfig()

			assert.Equal(t, tt.expectedEnv, Environment)
			assert.Equal(t, tt.expectedIntro, EnableIntrospection)
			assert.Equal(t, tt.expectedIsProd, IsProduction())
		})
	}
}

func TestExplicitIntrospectionOverrides(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedIntro bool
	}{
		{
			name: "enable introspection in production via ENABLE_INTROSPECTION=true",
			envVars: map[string]string{
				"ENVIRONMENT":          "production",
				"ENABLE_INTROSPECTION": "true",
			},
			expectedIntro: true,
		},
		{
			name: "enable introspection in production via ENABLE_INTROSPECTION=1",
			envVars: map[string]string{
				"ENVIRONMENT":          "production",
				"ENABLE_INTROSPECTION": "1",
			},
			expectedIntro: true,
		},
		{
			name: "enable introspection in production via INTROSPECTION_ENABLED=true",
			envVars: map[string]string{
				"ENVIRONMENT":           "production",
				"INTROSPECTION_ENABLED": "true",
			},
			expectedIntro: true,
		},
		{
			name: "disable introspection in development via ENABLE_INTROSPECTION=false",
			envVars: map[string]string{
				"ENVIRONMENT":          "development",
				"ENABLE_INTROSPECTION": "false",
			},
			expectedIntro: false,
		},
		{
			name: "disable introspection in development via ENABLE_INTROSPECTION=0",
			envVars: map[string]string{
				"ENVIRONMENT":          "development",
				"ENABLE_INTROSPECTION": "0",
			},
			expectedIntro: false,
		},
		{
			name: "disable introspection in development via INTROSPECTION_ENABLED=false",
			envVars: map[string]string{
				"ENVIRONMENT":           "development",
				"INTROSPECTION_ENABLED": "false",
			},
			expectedIntro: false,
		},
		{
			name: "ENABLE_INTROSPECTION takes precedence over INTROSPECTION_ENABLED",
			envVars: map[string]string{
				"ENVIRONMENT":           "production",
				"ENABLE_INTROSPECTION":  "false",
				"INTROSPECTION_ENABLED": "true",
			},
			expectedIntro: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv()
			defer clearConfigEnv()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			LoadConfig()

			assert.Equal(t, tt.expectedIntro, EnableIntrospection)
		})
	}
}
