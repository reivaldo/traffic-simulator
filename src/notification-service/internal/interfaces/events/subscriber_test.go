package eventsiface

import "testing"

func TestPackage(t *testing.T) {
	// JetStream integration is exercised in docker-compose; use case is covered in application tests.
	_ = NewSubscriber
}
