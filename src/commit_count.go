package main

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os/exec"
	"bufio"
	"os"
	"strings"
	"bytes"
	"strconv"
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

func fetchSource(repo Repository) error{
	fmt.Printf("Fetching %s\n", repo.Name)
	
	var cmd exec.Cmd = exec.Cmd {
		Path: "../bin/fetch-source",
		Args: []string {repo.Name, repo.Url},
	}
	
	err := cmd.Run()
	return err
}

func CountCommits(file_path string, repo_name string, setting Setting) map[string]int{
	var result map[string]int = make(map[string]int) 
	for _, contributor := range setting.Contributors {
		result[contributor.Name] = 0
	}
	
	inFile, _ := os.Open(file_path)
	defer inFile.Close()
	
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var line string = scanner.Text()
		for _, contributor := range setting.Contributors {
			if strings.Contains(line, contributor.Name) {
				result[contributor.Name]++
			}
		}
	}
	
	return result
} 

func CreateOutputFile(setting Setting, result map[string]map[string]int) {
	var buffer bytes.Buffer

	buffer.WriteString(",")
	for i, repo := range setting.Repositories {
		if i != 0 {
			buffer.WriteString(",")	
		}
		buffer.WriteString(repo.Name)
    }
	buffer.WriteString("\n")
	
	for _, contributor := range setting.Contributors {
		buffer.WriteString(contributor.Name)
		buffer.WriteString(",")
		for j, repo := range setting.Repositories {
			if j != 0 {
				buffer.WriteString(",")	
			}
			buffer.WriteString(strconv.Itoa(result[contributor.Name][repo.Name]))
		}
		buffer.WriteString("\n")
	}
	fmt.Printf(buffer.String())
	ioutil.WriteFile("work/result.csv", buffer.Bytes(), 0644)

}

func main() {
	fmt.Printf("Reading Setting File: setting.yml\n")
	setting, err := ReadSettingFile("setting.yml")
	if err != nil { panic(err) }
	
	var result map[string]map[string]int = make(map[string]map[string]int)
	for _, contributor := range setting.Contributors {
		result[contributor.Name] = make(map[string]int)
	} 
	
	fmt.Printf("Fetching History\n")
	for _, repo := range setting.Repositories {
		fetchSource(repo)
		
		var file_path string = "work/" + repo.Name + "_log.txt" 
		var repo_counts map[string]int = CountCommits(file_path, repo.Name, setting)
		
		for _, contributor := range setting.Contributors {
			result[contributor.Name][repo.Name] = repo_counts[contributor.Name]
		}
	}
	
	CreateOutputFile(setting, result)
			
}
