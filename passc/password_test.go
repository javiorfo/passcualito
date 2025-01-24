package passc

import "testing"

func TestAlignPassword(t *testing.T) {
	result := alignPassword("12345678910121415")
	if len(result) != 16 {
		t.Fatal("length must be 16")
	}

	result = alignPassword("123421415")
	if len(result) != 16 {
		t.Fatal("length must be 16")
	}
}
