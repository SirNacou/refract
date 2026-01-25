package idgen

import "testing"

func TestPrintEncodings(t *testing.T) {
	// t.Skip("Helper test - run manually to see encodings")

	tests := []uint64{
		0,
		123,
		1234567890,
		1234567890123456,
		^uint64(0),
	}

	for _, id := range tests {
		encoded := EncodeBase62(id)
		t.Logf("ID: %d -> Base62: %s", id, encoded)
	}
}
