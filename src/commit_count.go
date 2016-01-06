package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
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

func fetchSource(repo Repository) error {
	fmt.Printf("Fetching %s (%s)\n", repo.Name, repo.Url)

	var cmd *exec.Cmd = exec.Command(
		"bin/fetch-source",
		repo.Name,
		repo.Url)

	_, err := cmd.CombinedOutput()
	return err
}

type GitCommit struct {
	Author         string
	Date           time.Time
	Description    string
	CoAuthor       string
	Repo           string
	AuthorDomain   string
	CoAuthorDomain string
}

func GetCoAuthor(line string) string {
	elements := strings.Split(strings.Trim(line, " "), " ")
	if len(elements) >= 3 {
		result := elements[1] + " " + elements[2]
		return result
	} else {
		// fmt.Printf("GetCoAuthor: %s\n", line)
		return ""
	}
}

func GetEmailDomain(line string) string {
	// fmt.Printf("line = %s\n", line)
	elements := strings.Split(strings.Trim(line, " "), " ")
	var resultElements []string = strings.Split(elements[len(elements)-1], "@")
	if len(resultElements) != 2 {
		// fmt.Printf("line = %s\n", line)
		return ""
	} else {
		var result string = resultElements[1]
		return result[0 : len(result)-1]
	}
}

func GetFirstWord(line string) string {
	elements := strings.Split(strings.Trim(line, " "), " ")
	return elements[0]
}

func parseDate(line string) time.Time {
	var elements []string = strings.Split(line, " ")
	var dateString string = elements[7] + "-" + elements[4] + "-" + elements[5]
	result, _ := time.Parse("2006-Jan-02", dateString)
	// fmt.Printf("Date: %s\n", dateString)
	return result
}

func ReadCommit(scanner *bufio.Scanner, repo string) []GitCommit {
	var result []GitCommit

	for scanner.Scan() {
		var line string = scanner.Text()

		// var pair string

		var firstWord string = GetFirstWord(line)

		if firstWord == "Author:" {

			var author string
			var description string
			var coauthor string
			var authorDomain string
			var coauthorDomain string
			var date time.Time

			isTwoAuthorPattern, author, coauthor := IsTwoAuthorPattern(line)
			if !isTwoAuthorPattern {
				author = GetAuthor(line)
				authorDomain = GetEmailDomain(line)
			}

			// Date
			scanner.Scan()
			date = parseDate(scanner.Text())

			// Blank line
			scanner.Scan()
			// scanner.Text()

			for scanner.Scan() {
				line = scanner.Text()

				firstWord = GetFirstWord(line)

				if firstWord == "Signed-off-by:" {
					coauthor = GetCoAuthor(line)
					coauthorDomain = GetEmailDomain(line)

					break
				} else if firstWord == "commit" {
					break
				}

				description += " "
				description += strings.Trim(line, " ")
			}

			description = strings.Trim(description, " ")

			commit := GitCommit{
				Author:         author,
				Date:           date,
				Description:    description,
				CoAuthor:       coauthor,
				Repo:           repo,
				AuthorDomain:   authorDomain,
				CoAuthorDomain: coauthorDomain,
			}
			result = append(result, commit)

		}
	}

	return result
}

func GetAuthor(line string) string {
	elements := strings.Split(line, " ")
	result := elements[1] + " " + elements[2]
	return result
}

func IsTwoAuthorPattern(line string) (bool, string, string) {
	if strings.Contains(line, " and ") {
		elements := strings.Split(line, " ")
		if elements[3] == "and" {
			author1 := elements[1] + " " + elements[2]
			author2 := elements[4] + " " + elements[5]
			return true, author1, author2
		}
	}
	return false, "", ""
}

func CountCommits(file_path string, repo_name string, setting Setting, log_buffer bytes.Buffer) map[string]int {
	fmt.Printf("Reading log file: %s\n", file_path)

	var result map[string]int = make(map[string]int)
	for _, contributor := range setting.Contributors {
		result[contributor.Name] = 0
	}

	inFile, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
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
			fmt.Printf("%s at repo %s = %d\n", contributor.Name, repo.Name, result[contributor.Name][repo.Name])
			buffer.WriteString(strconv.Itoa(result[contributor.Name][repo.Name]))
		}
		buffer.WriteString("\n")
	}
	fmt.Printf(buffer.String())
	ioutil.WriteFile("work/result.csv", buffer.Bytes(), 0644)

}

func IsEmcCommit(commit GitCommit, contributors []Contributor) (bool, string) {
	for _, contributor := range contributors {
		if contributor.Name == commit.Author {
			return true, contributor.Name
		}
		if contributor.Name == commit.CoAuthor {
			return true, contributor.Name
		}
	}

	return false, ""
}

func CreateLogOutputFile(setting Setting, log_result map[string][]GitCommit) {
	var buffer bytes.Buffer

	buffer.WriteString("Author,CoAuthor,Code Repo,Commit Description\n")
	for _, contributor := range setting.Contributors {
		var gitCommits []GitCommit = log_result[contributor.Name]
		for _, commit := range gitCommits {
			buffer.WriteString(commit.Author)
			buffer.WriteString(",")
			buffer.WriteString(commit.CoAuthor)
			buffer.WriteString(",")
			buffer.WriteString(commit.Repo)
			buffer.WriteString(",")
			buffer.WriteString(commit.Description)
			buffer.WriteString("\n")
		}
	}

	fmt.Printf(buffer.String())
	ioutil.WriteFile("work/result_log.csv", buffer.Bytes(), 0644)
}

func main() {
	fmt.Printf("Reading Setting File: setting.yml\n")
	setting, err := ReadSettingFile("setting.yml")
	if err != nil {
		panic(err)
	}

	var count_result map[string]map[string]int = make(map[string]map[string]int)
	var log_result map[string][]GitCommit = make(map[string][]GitCommit)

	for _, contributor := range setting.Contributors {
		count_result[contributor.Name] = make(map[string]int)
		log_result[contributor.Name] = make([]GitCommit, 0)
	}

	fmt.Printf("Fetching History\n")
	var wg sync.WaitGroup
	// var log_buffer bytes.Buffer
	for _, repo := range setting.Repositories {
		wg.Add(1)
		go func(repo1 Repository) {
			defer wg.Done()

			fetch_error := fetchSource(repo1)
			if fetch_error != nil {
				panic(fetch_error)
			}

			var file_path string = "work/" + repo1.Name + "_log.txt"
			inFile, err := os.Open(file_path)
			if err != nil {
				panic(err)
			}
			defer inFile.Close()

			scanner := bufio.NewScanner(inFile)
			var commits []GitCommit = ReadCommit(scanner, repo1.Name)
			for _, commit := range commits {
				// When Author and CoAuthor are both EMC, only counts as 1
				isEmcCommit, contributorName := IsEmcCommit(commit, setting.Contributors)
				if isEmcCommit {
					count_result[contributorName][repo1.Name] += 1
					log_result[contributorName] = append(log_result[contributorName], commit)
				}
			}

			// var repo_counts map[string]int = CountCommits(file_path, repo1.Name, setting, log_buffer)
			//
			// for _, contributor := range setting.Contributors {
			// 	result[contributor.Name][repo1.Name] = repo_counts[contributor.Name]
			// }
		}(repo)
	}

	wg.Wait()
	CreateLogOutputFile(setting, log_result)
	CreateOutputFile(setting, count_result)

	FetchOverallCount()

}

func getRepoName(url string) string {
	elements := strings.Split(url, "/")
	return strings.Split(elements[len(elements)-1], ".")[0]
}

func getScanner(repoName string) *bufio.Scanner {
	var file_path string = "work/" + repoName + "_log.txt"
	inFile, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	// defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	return scanner
}

func CountOverallCommit(gitCommits []GitCommit, result map[string]int, beginDate time.Time) {
	for _, commit := range gitCommits {

		if commit.Date.After(beginDate) {
			if commit.AuthorDomain != "" {
				result[commit.AuthorDomain]++
				result["TOTAL"]++
			}

			if commit.CoAuthorDomain != "" {
				result[commit.CoAuthorDomain]++
				result["TOTAL"]++
			}
		}
	}
}

func getRepos(filepath string) map[string]string {
	var result map[string]string = make(map[string]string)

	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var url string = scanner.Text()
		var name = getRepoName(url)
		result[name] = url
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return result
}

func getDate(dateString string) time.Time {
	result, err := time.Parse("2006-01-02", dateString)

	if err != nil {
		panic(err)
	}

	return result
}

func FetchOverallCount() {
	var repoMap map[string]string = getRepos("repos.txt")
	var result map[string]int = make(map[string]int)

	var beginDate time.Time = getDate("2015-05-31")

	concurrency := 30
	sem := make(chan bool, concurrency)

	var wg1 sync.WaitGroup
	for repoName, url := range repoMap {

		var repo Repository = Repository{Name: repoName, Url: url}
		wg1.Add(1)
		sem <- true
		go func(repo1 Repository) {
			defer wg1.Done()
			defer func() { <-sem }()
			fetch_error := fetchSource(repo1)

			if fetch_error != nil {
				fmt.Printf("ERROR FETCH: %s\n", repo.Name)
				panic(fetch_error)
			}

			var scanner *bufio.Scanner = getScanner(repo1.Name)
			var gitCommits []GitCommit = ReadCommit(scanner, repo1.Name)

			CountOverallCommit(gitCommits, result, beginDate)
			fmt.Printf("COUNT = %d, TOTAL = %d (%s)\n", len(result), result["TOTAL"], repo1.Name)
		}(repo)
	}

	fmt.Printf("Waiting for Count to finish...\n")
	wg1.Wait()

	fmt.Printf("Generating Output\n")
	for k, v := range result {
		fmt.Printf("%s = %d\n", k, v)
	}

	CreateTotalCountOutputFile(result)
}

func CreateTotalCountOutputFile(result map[string]int) {
	var buffer bytes.Buffer

	for k, v := range result {
		buffer.WriteString(k)
		buffer.WriteString(",")
		buffer.WriteString(strconv.Itoa(v))
		buffer.WriteString("\n")
	}

	ioutil.WriteFile("work/total_count.csv", buffer.Bytes(), 0644)
}
