package postgresinfra

import "testing"

func TestPackage(t *testing.T) {
	// Repository needs a real *pgxpool.Pool; behaviour is validated via processor application tests + integration.
	_ = NewMessageRepository
}
