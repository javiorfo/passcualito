package passc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/javiorfo/steams"
)

type Data struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Info     string `json:"info"`
}

func (d Data) toJSON() (*string, error) {
	json, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	res := string(json)
	return &res, nil
}

func (d *Data) fromJSON(jsonStr []byte) error {
	err := json.Unmarshal(jsonStr, d)
	if err != nil {
		return err
	}
	return nil
}

func (d Data) print(isEnd bool) {
	fmt.Println("│")
	fmt.Println("├── \033[1mname:\033[0m    ", d.Name)
	fmt.Println("├── \033[1mpassword:\033[0m", d.Password)
	if isEnd {
		fmt.Println("└── \033[1minfo:\033[0m    ", d.Info)
	} else {
		fmt.Println("├── \033[1minfo:\033[0m    ", d.Info)
	}
}

func stringToDataSlice(content string) []Data {
	items := strings.Split(content, passcItemSeparator)
	return steams.Mapping(steams.OfSlice(items), func(v string) Data {
		var data Data
		err := data.fromJSON([]byte(v))
		if err != nil {
			// TODO replace with log
			fmt.Println(err.Error())
		}
		return data
	}).Collect()
}
