package passc

import "testing"

func TestData(t *testing.T) {
	data := Data{
		Name:      "github",
		Password: "12345",
	}
	json, err := data.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Data to JSON: %s", *json)

	jsonStr := `{
        "key": "jsonkey", 
        "password": "23412342", 
        "info": "some info"
    }`

	data.FromJSON([]byte(jsonStr))
	t.Logf("Data from JSON: %#v", data)
}
