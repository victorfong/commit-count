package main

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"exec"
)

type Repository struct {
	Name string
	Url  string
}

type Contributor struct {
	Name string
}

type Setting struct {
	Repositories []Repository
	Contributors []Contributor
}

func UnmarshalYaml(data []byte) (Setting, error) {
	t := Setting{}

	err := yaml.Unmarshal(data, &t)
	if err != nil {
		return Setting{}, err
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	return t, nil
}

func ReadFile(file_path string) ([]byte, error) {
	dat, err := ioutil.ReadFile(file_path)
	return dat, err
}

func ReadSettingFile(file_path string) (Setting, error) {
	dat, err := ioutil.ReadFile(file_path)
	if err != nil {
		return Setting{}, err
	}
	
	setting, err := UnmarshalYaml(dat)
	return setting, err
}



//func fetchSource(repo Repository) {
//	exec.Run(app, []string{app, "-l"}, nil, "", exec.DevNull, exec.Pipe, exec.Pipe)
//}

func main() {
	fmt.Printf("Reading Setting File: setting.yml\n")
	setting, err := ReadSettingFile("setting.yml")
	if err != nil { panic(err) }
	
	
	
	
	
	
	
	
	
	
}
