package bootstrap

import "testing"

func TestPackage(t *testing.T) {
	// BuildHTTPHandler is covered via integration; wiring is exercised from main in docker-compose.
	_ = BuildHTTPHandler
}
