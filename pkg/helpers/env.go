package helpers

import (
	"os"
	"strconv"
)

func GetEnv(envVarName string, defaultValue string) string {
	value := os.Getenv(envVarName)
	if value != "" {
		return value
	}
	return defaultValue
}

func GetIntEnv(envVarName string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(envVarName))
	if err != nil {
		value = defaultValue
	}
	return value
}
