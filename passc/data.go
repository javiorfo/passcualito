package passc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/javiorfo/steams"
)

type Data struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Info     string `json:"info"`
}

func newData(name, password, info string) (*Data, error) {
	data := Data{
		Name: name,
		Info: info,
	}

	if password == "" {
		pwd, err := generateRandomPasswordDefault()
		if err != nil {
			return nil, err
		}
		password = *pwd
	}
	data.Password = password
	return &data, nil
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
		_ = err // Unimplemented
		return data
	}).Collect()
}

func isNameTaken(content, name string) bool {
	items := strings.Split(content, passcItemSeparator)
	return steams.OfSlice(items).AnyMatch(func(v string) bool {
		var data Data
		err := data.fromJSON([]byte(v))
		_ = err
		return data.Name == name
	})
}

func getDataSliceFromJsonFile(filePath string) ([]Data, error) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var data []Data
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	if err := getRepeatedNames(data); err != nil {
		return nil, err
	}

	return data, nil
}

func getRepeatedNames(dataSlice []Data) error {
	repeated := steams.GroupByCounting(steams.OfSlice(dataSlice), func(d Data) string {
		return d.Name
	}).Filter(func(s string, i int) bool {
		return i > 1
	}).KeysToSteam().Collect()

	if len(repeated) > 0 {
		return fmt.Errorf("repeated names (%s)", strings.Join(repeated, ", "))
	}
	return nil
}
