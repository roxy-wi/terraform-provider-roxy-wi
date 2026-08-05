package roxywi

import "testing"

func TestProviderInternalValidate(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema validation failed: %v", err)
	}
}
