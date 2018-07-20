package main

import (
	"os"
	"strconv"
	"strings"
)

// CASTemplateFeatureGate returns true if cas template feature gate is
// enabled
func CASTemplateFeatureGate() (bool, error) {
	return strconv.ParseBool(lookEnv(CASTemplateFeatureGateENVK))
}

// lookENV wrapper over LookupEnv to retrieves the value of the environment
// variable named by the key.
func lookEnv(envKey string) string {
	val, present := os.LookupEnv(string(envKey))
	if !present {
		return "false"
	}
	return strings.TrimSpace(val)
}
