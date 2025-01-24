package passc

import "encoding/json"

type Data struct {
	Key      string  `json:"key"`
	Password string  `json:"password"`
	Info     *string `json:"info,omitempty"`
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
