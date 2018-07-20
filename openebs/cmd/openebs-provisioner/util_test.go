package main

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestCASTemplateFeatureGate(t *testing.T) {

	cases := map[string]struct {
		key, value  string
		expectValue bool
		expectErr   error
	}{
		"Incorrect value on": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "on",
			expectValue: false,
			expectErr:   errors.New("invalid syntax"),
		},
		"Key and value nil": {
			key:         "",
			value:       "",
			expectValue: false,
			expectErr:   nil,
		},
		"Value is nil": {
			key:         CASTemplateFeatureGateENVK,
			value:       "",
			expectValue: false,
			expectErr:   errors.New("invalid syntax"),
		},
		"Valid key and value": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "true",
			expectValue: true,
			expectErr:   nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			os.Setenv(tc.key, tc.value)
			defer os.Unsetenv(tc.key)

			feature, err := CASTemplateFeatureGate()
			if tc.expectErr != nil {
				if !reflect.DeepEqual(tc.expectErr, err.(*strconv.NumError).Err) {
					t.Errorf("Expected %s, got %s", tc.expectErr, err)
				}
			}
			if !reflect.DeepEqual(feature, tc.expectValue) {
				t.Errorf("Expected %v, got %v", tc.expectValue, feature)
			}
		})
	}
}
