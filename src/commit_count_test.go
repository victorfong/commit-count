package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "strings"
    "bufio"
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
  assert.Equal(t, "", GetFirstWord(" "))
  assert.Equal(t, "", GetFirstWord(""))
}

var test_commit = `
Author: Maria Shaldibina <mshaldibina@pivotal.io>
Date:   Thu Oct 15 09:43:35 2015 -0700

    Merge branch 'master' into hotfix-postgres

    Signed-off-by: Tyler Schultz <tschultz@pivotal.io>

commit 7ce9e8b628034446c28b4955863386fbf4aa8c1d
Author: Devin Fallak <dfallak@pivotal.io>
Date:   Wed Oct 14 15:44:34 2015 -0400

    Update README.md


`

func TestReadCommit(t *testing.T){
  scanner := bufio.NewScanner(strings.NewReader(test_commit))
  var gitCommits []GitCommit = ReadCommit(scanner, "repo1")
  assert.Equal(t, 2, len(gitCommits))

  assert.Equal(t, "Maria Shaldibina", gitCommits[0].Author)
  assert.Equal(t, "Merge branch 'master' into hotfix-postgres", gitCommits[0].Description)
  assert.Equal(t, "repo1", gitCommits[0].Repo)

  assert.Equal(t, "Devin Fallak", gitCommits[1].Author)
  assert.Equal(t, "Update README.md", gitCommits[1].Description)
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
