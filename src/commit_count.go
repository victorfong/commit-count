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
	
	var cmd *exec.Cmd = exec.Command(
		"bin/fetch-source",
		repo.Name, 
		repo.Url)
	
	result, err := cmd.CombinedOutput()
	fmt.Printf("result = " + string(result))
	
	return err
}

func CountCommits(file_path string, repo_name string, setting Setting) map[string]int{
	fmt.Printf("Reading log file: %s\n", file_path)
	
	var result map[string]int = make(map[string]int) 
	for _, contributor := range setting.Contributors {
		result[contributor.Name] = 0
	}
	
	inFile, err := os.Open(file_path)
	if err != nil { panic(err) }
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
	ioutil.WriteFile("~/tmp/commit-count/work/result.csv", buffer.Bytes(), 0644)

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
		fetch_error := fetchSource(repo)
		if fetch_error != nil {
			panic(fetch_error)
		}
		
		var file_path string = "/home/victor/tmp/commit-count/work/" + repo.Name + "_log.txt" 
		var repo_counts map[string]int = CountCommits(file_path, repo.Name, setting)
		
		for _, contributor := range setting.Contributors {
			result[contributor.Name][repo.Name] = repo_counts[contributor.Name]
		}
	}
	
	CreateOutputFile(setting, result)
			
}
