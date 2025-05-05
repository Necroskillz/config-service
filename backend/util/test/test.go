package test

import "testing"

func RunCases[TC any](t *testing.T, run func(*testing.T, TC), testCases map[string]TC) {
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
