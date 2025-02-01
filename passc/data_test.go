package passc

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestData(t *testing.T) {
	data := Data{
		Name:     "github",
		Password: "12345",
	}
	json, err := data.toJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Data to JSON: %s", *json)

	jsonStr := `{
        "name": "jsonkey", 
        "password": "23412342", 
        "info": "some info"
    }`

	data.fromJSON([]byte(jsonStr))
	t.Logf("Data from JSON: %#v", data)
}

func TestGetDataSliceFromJsonFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_data.json")
	if err != nil {
		t.Errorf("Failed to create temporary file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	sampleData := []Data{
		{Name: "John", Password: "30"},
		{Name: "Jane", Password: "25"},
	}

	jsonData, err := json.Marshal(sampleData)
	if err != nil {
		t.Errorf("Failed to marshal sample data: %v", err)
		return
	}

	_, err = tempFile.Write(jsonData)
	if err != nil {
		t.Errorf("Failed to write to temporary file: %v", err)
		return
	}

	data, err := getDataSliceFromJsonFile(tempFile.Name())
	if err != nil {
		t.Errorf("getDataSliceFromJsonFile returned an error: %v", err)
		return
	}

	if !reflect.DeepEqual(data, sampleData) {
		t.Errorf("Returned data does not match sample data:\nExpected: %v\nGot: %v", sampleData, data)
	}
}

func TestStringToDataSlice(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Data
	}{
		{
			name:    "single item",
			content: `{"name":"John","password":"other value"}`,
			expected: []Data{
				{Name: "John", Password: "other value"},
			},
		},
		{
			name:    "multiple items",
			content: `{"name":"John","password":"other value"}|{"name":"Jane","password":"another value"}`,
			expected: []Data{
				{Name: "John", Password: "other value"},
				{Name: "Jane", Password: "another value"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := stringToDataSlice(test.content)
			if len(actual) != len(test.expected) {
				t.Errorf("%s: expected %d items, got %d", test.name, len(test.expected), len(actual))
			}
			for i, item := range actual {
				if item.Name != test.expected[i].Name || item.Password != test.expected[i].Password {
					t.Errorf("expected item %d to be %+v, got %+v", i, test.expected[i], item)
				}
			}
		})
	}
}

func TestIsNameTaken(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		nameInput string
		expected  bool
	}{
		{
			name:      "name taken",
			content:   `{"name":"John","password":"other value"}|{"name":"Jane","password":"another value"}`,
			nameInput: "John",
			expected:  true,
		},
		{
			name:      "name not taken",
			content:   `{"name":"John","password":"other value"}|{"name":"Jane","password":"another value"}`,
			nameInput: "Bob",
			expected:  false,
		},
		{
			name:      "empty content",
			content:   "",
			nameInput: "John",
			expected:  false,
		},
		{
			name:      "invalid JSON",
			content:   `{"some":"John"}`,
			nameInput: "John",
			expected:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := isNameTaken(test.content, test.nameInput)
			if actual != test.expected {
				t.Errorf("expected %t, got %t", test.expected, actual)
			}
		})
	}
}

func TestGetRepeatedNames(t *testing.T) {
	tests := []struct {
		name      string
		dataSlice []Data
		expected  error
	}{
		{
			name: "no repeated names",
			dataSlice: []Data{
				{Name: "John"},
				{Name: "Jane"},
				{Name: "Bob"},
			},
			expected: nil,
		},
		{
			name: "single repeated name",
			dataSlice: []Data{
				{Name: "John"},
				{Name: "John"},
				{Name: "Bob"},
			},
			expected: fmt.Errorf("repeated names (John)"),
		},
		{
			name:      "empty data slice",
			dataSlice: []Data{},
			expected:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getRepeatedNames(test.dataSlice)
			if actual != nil && test.expected != nil {
				if actual.Error() != test.expected.Error() {
					t.Errorf("expected error %q, got %q", test.expected, actual)
				}
			} else if actual != nil || test.expected != nil {
				t.Errorf("expected error %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestIsNameMatched(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected bool
	}{
		{
			name: "soma",
			data: Data{
				Name: "some",
			},
			expected: false,
		},
		{
			name: "som",
			data: Data{
				Name: "some",
			},
			expected: true,
		},
		{
			name: "s",
			data: Data{
				Name: "some",
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.data.isNameMatched(test.name)
			if result != test.expected {
				t.Errorf("expected %t, got %t", test.expected, result)
			}
		})
	}
}
