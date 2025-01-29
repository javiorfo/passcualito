package passc

import (
	"encoding/json"
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

func TestGetRepeatedNames(t *testing.T) {
	const values = `[
        {
            "name": "javi",
            "password": "123",
            "info": "some"
        },
        {
            "name": "javi",
        },
        {
            "name": "javi2",
            "password": "1234"
        }
    ]`
}
