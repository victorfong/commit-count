package main

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var test_data = `
---
repositories:
- name: Bosh
  url: some_url
contributors:
- name: Victor Fong
`

func TestReadSetting(t *testing.T) {
	result, err := UnmarshalYaml([]byte(test_data))

	assert.Equal(t, nil, err)

	assert.Equal(t, 1, len(result.Repositories))
	assert.Equal(t, "Bosh", result.Repositories[0].Name)
	assert.Equal(t, "some_url", result.Repositories[0].Url)

	assert.Equal(t, 1, len(result.Contributors))
	assert.Equal(t, "Victor Fong", result.Contributors[0].Name)
}

func TestReadSettingFile(t *testing.T) {
	result, err := ReadSettingFile("test_setting.yml")

	assert.Equal(t, nil, err)

	assert.Equal(t, 2, len(result.Repositories))
	assert.Equal(t, "Bosh", result.Repositories[0].Name)
	assert.Equal(t, "some_url", result.Repositories[0].Url)

	assert.Equal(t, 2, len(result.Contributors))
	assert.Equal(t, "Victor Fong", result.Contributors[0].Name)
}

func TestGetFirstWord(t *testing.T) {

	assert.Equal(t, "Authors:", GetFirstWord("Authors: Victor Fong"))
	assert.Equal(t, "Date:", GetFirstWord("Date:   Sun Oct 18 19:33:09 2015 -0400"))
	assert.Equal(t, "", GetFirstWord(" "))
	assert.Equal(t, "", GetFirstWord(""))

}

var test_commit = `
commit a39b69d7e6ab6c59c76102136815c6b7ae578804
Author: Maria Shaldibina <mshaldibina@pivotal.io>
Date:   Thu Oct 15 09:43:35 2015 -0700

    Merge branch 'master' into hotfix-postgres

    Signed-off-by: Tyler Schultz <tschultz@pivotal.io>

commit 7ce9e8b628034446c28b4955863386fbf4aa8c1d
Author: Devin Fallak <dfallak@pivotal.io>
Date:   Wed Oct 14 15:44:34 2015 -0400

    Update README.md


`

func TestReadCommit(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(test_commit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 2, len(gitCommits))

	assert.Equal(t, "Maria Shaldibina", gitCommits[0].Author)
	assert.Equal(t, "Merge branch 'master' into hotfix-postgres", gitCommits[0].Description)
	assert.Equal(t, "repo1", gitCommits[0].Repo)
	var date time.Time = getDate("2015-10-15")
	assert.True(t, date.Equal(gitCommits[0].Date))

	assert.Equal(t, "Devin Fallak", gitCommits[1].Author)
	assert.Equal(t, "Update README.md", gitCommits[1].Description)
	var date2 time.Time = getDate("2015-10-14")
	assert.True(t, date2.Equal(gitCommits[1].Date))

}

func TestGetCoAuthor(t *testing.T) {
	var testString = "    Signed-off-by: Tyler Schultz <tschultz@pivotal.io>"
	assert.Equal(t, "Tyler Schultz", GetCoAuthor(testString))
}

func TestGetAuthor(t *testing.T) {
	var testString = "Author: Devin Fallak <dfallak@pivotal.io>"
	assert.Equal(t, "Devin Fallak", GetAuthor(testString))
}

var test_commit2 = `
Author: Victor Fong <victor.fong@emc.com>
Date:   Thu Oct 15 09:43:35 2015 -0700

    Merge branch 'master' into hotfix-postgres

    Signed-off-by: Tyler Schultz <tschultz@pivotal.io>
`

func TestIsEmcCommit(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(test_commit2))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 1, len(gitCommits))

	setting, _ := UnmarshalYaml([]byte(test_data))
	isEmcCommit, name := IsEmcCommit(gitCommits[0], setting.Contributors)

	assert.Equal(t, true, isEmcCommit)
	assert.Equal(t, "Victor Fong", name)

}

func TestIsTwoAuthorPattern(t *testing.T) {
	var testString = "Author: Chris Piraino and Yu Zhang <cpiraino@pivotal.io>"
	result, author1, author2 := IsTwoAuthorPattern(testString)

	assert.Equal(t, true, result)
	assert.Equal(t, "Chris Piraino", author1)
	assert.Equal(t, "Yu Zhang", author2)
}

func TestGetRepoName(t *testing.T) {
	var testString string = "https://github.com/cloudfoundry/nodejs-buildpack.git"
	var result string = getRepoName(testString)
	assert.Equal(t, "nodejs-buildpack", result)
}

func TestGetCoAuthorDomain(t *testing.T) {
	var testString string = "    Signed-off-by: Min Su Han <glide1@gmail.com>"
	var result string = GetEmailDomain(testString)
	assert.Equal(t, "gmail.com", result)
}

func TestGetAuthorDomain(t *testing.T) {
	var testString string = "Author: Maria Shaldibina <mshaldibina@pivotal.io>"
	var result string = GetEmailDomain(testString)
	assert.Equal(t, "pivotal.io", result)
}

func TestGetTestDomain(t *testing.T) {
	var testString string = "line = Author: test <test>"
	var result string = GetEmailDomain(testString)
	assert.Equal(t, "", result)
}

func TestParseDate(t *testing.T) {
	var testString string = "Date:   Sun Oct 18 17:44:34 2015 -0400"
	var result time.Time = parseDate(testString)
	expectedResult, _ := time.Parse("01-02-2006", "10-18-2015")
	assert.True(t, expectedResult.Equal(result))
}

func TestGetDate(t *testing.T) {
	var testString string = "Date:   Sun Oct 18 17:44:34 2015 -0400"
	var result time.Time = parseDate(testString)

	var date time.Time = getDate("2015-05-31")
	assert.True(t, result.After(date))
}

var testCommit = `
commit a39b69d7e6ab6c59c76102136815c6b7ae578804
Author: Marco Voelz <marco.voelz@sap.com>
Date:   Tue Dec 29 17:56:22 2015 +0100

    Add instructions to run tests

commit 4d4033620e0c7280c8354504358a17b510c32e3f
Author: Marco Voelz <marco.voelz@sap.com>
Date:   Mon Dec 28 17:02:41 2015 +0100

    Remove space in 'new final release' commit msg

commit d89a0dc09f0a9948e02cc47220e0db2967e3cc7e
Author: Marco Voelz <marco.voelz@sap.com>
Date:   Mon Dec 28 16:56:56 2015 +0100

    Final releases are built in concourse

commit 078744d4ccfd72f198dd15c210e689cc6929201b
Merge: 3c71e67 0a09bc0
Author: Beyhan Veli <beyhan.veli@sap.com>
Date:   Tue Dec 29 15:22:50 2015 +0100

    Merge pull request #17 from hashmap/power-builder

    Enable ppc64le support

commit 3c71e67c27ba0f4232b004e13b1fe6486b7b945b
Author: Beyhan Veli <beyhan.veli@sap.com>
Date:   Tue Dec 22 14:01:09 2015 +0100

    Add unit tests for key_name configuration

    - key_name can be configured in resource_pool and
      CPI properties. Unit tests added to test this feature.
    - ITs configure key_name only as CPI property

    [#108602248](https://www.pivotaltracker.com/story/show/108602248)

    Signed-off-by: Felix Riegger <felix.riegger@sap.com>
`

func TestCountOverallCommit_None(t *testing.T){
	scanner := bufio.NewScanner(strings.NewReader(testCommit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 5, len(gitCommits))

	var beginDate time.Time = getDate("2015-01-23")
	var endDate time.Time = getDate("2016-01-01")
	var result map[string]int = make(map[string]int)

	CountOverallCommit(gitCommits, result, beginDate, endDate)
	assert.Equal(t, 6, result["TOTAL"])
	assert.Equal(t, 6, result["sap.com"])
}

func TestCountOverallCommit_IgnoreOne(t *testing.T){
	scanner := bufio.NewScanner(strings.NewReader(testCommit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 5, len(gitCommits))

	var beginDate time.Time = getDate("2015-12-23")
	var endDate time.Time = getDate("2016-01-01")
	var result map[string]int = make(map[string]int)

	CountOverallCommit(gitCommits, result, beginDate, endDate)
	assert.Equal(t, 4, result["TOTAL"])
	assert.Equal(t, 4, result["sap.com"])
}

func TestCountOverallCommit_IgnoreThree(t *testing.T){
	scanner := bufio.NewScanner(strings.NewReader(testCommit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 5, len(gitCommits))

	var beginDate time.Time = getDate("2015-12-28")
	var endDate time.Time = getDate("2016-01-01")
	var result map[string]int = make(map[string]int)

	CountOverallCommit(gitCommits, result, beginDate, endDate)
	assert.Equal(t, 2, result["TOTAL"])
	assert.Equal(t, 2, result["sap.com"])
}

func TestCountOverallCommit_IgnoreEndTwo(t *testing.T){
	scanner := bufio.NewScanner(strings.NewReader(testCommit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 5, len(gitCommits))

	var beginDate time.Time = getDate("2015-01-23")
	var endDate time.Time = getDate("2015-12-29")
	var result map[string]int = make(map[string]int)

	CountOverallCommit(gitCommits, result, beginDate, endDate)
	assert.Equal(t, 4, result["TOTAL"])
	assert.Equal(t, 4, result["sap.com"])
}

func TestCountOverallCommit_IgnoreAll(t *testing.T){
	scanner := bufio.NewScanner(strings.NewReader(testCommit))
	var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
	assert.Equal(t, 5, len(gitCommits))

	var beginDate time.Time = getDate("2015-12-23")
	var endDate time.Time = getDate("2015-12-25")
	var result map[string]int = make(map[string]int)

	CountOverallCommit(gitCommits, result, beginDate, endDate)
	assert.Equal(t, 0, result["TOTAL"])
	assert.Equal(t, 0, result["sap.com"])
}



func TestGetRepos(t *testing.T) {
	var result map[string]string = getRepos("test_repo.txt")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "https://github.com/cloudfoundry/api-docs.git", result["api-docs"])
	assert.Equal(t, "https://github.com/cloudfoundry/binary-builder.git", result["binary-builder"])
}
