package passc

import (
	"testing"

	"github.com/javiorfo/nilo"
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
		size     nilo.Option[int]
		charset  nilo.Option[string]
		expected int
	}{
		{
			name:     "default size and charset",
			size:     nilo.Nil[int](),
			charset:  nilo.Nil[string](),
			expected: 20,
		},
		{
			name:     "custom size",
			size:     nilo.Value(10),
			charset:  nilo.Nil[string](),
			expected: 10,
		},
		{
			name:     "custom charset",
			size:     nilo.Nil[int](),
			charset:  nilo.Value("abcdefghijklmnopqrstuvwxyz"),
			expected: 20,
		},
		{
			name:     "custom size and charset",
			size:     nilo.Value(15),
			charset:  nilo.Value("abcdefghijklmnopqrstuvwxyz"),
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
