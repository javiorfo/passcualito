package passc

import (
	"testing"

	"github.com/javiorfo/steams/opt"
)

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

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name     string
		size     opt.Optional[int]
		charset  opt.Optional[string]
		expected int
	}{
		{
			name:     "default size and charset",
			size:     opt.Empty[int](),
			charset:  opt.Empty[string](),
			expected: 20,
		},
		{
			name:     "custom size",
			size:     opt.Of(10),
			charset:  opt.Empty[string](),
			expected: 10,
		},
		{
			name:     "custom charset",
			size:     opt.Empty[int](),
			charset:  opt.Of("abcdefghijklmnopqrstuvwxyz"),
			expected: 20,
		},
		{
			name:     "custom size and charset",
			size:     opt.Of(15),
			charset:  opt.Of("abcdefghijklmnopqrstuvwxyz"),
			expected: 15,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			password, err := generateRandomPassword(test.size, test.charset)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if len(*password) != test.expected {
				t.Errorf("expected password length %d, got %d", test.expected, len(*password))
			}
		})
	}
}
