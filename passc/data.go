package passc

import "encoding/json"

type Data struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Info     string `json:"info"`
}

func (d Data) ToJSON() (*string, error) {
	json, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	res := string(json)
	return &res, nil
}

func (d *Data) FromJSON(jsonStr []byte) error {
	err := json.Unmarshal(jsonStr, d)
	if err != nil {
		return err
	}
	return nil
}
