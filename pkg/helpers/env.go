package helpers

import "os"

func GetEnv(envVarName string, defaultValue string) string {
	value := os.Getenv(envVarName)
	if value != "" {
		return value
	}
	return defaultValue
}
