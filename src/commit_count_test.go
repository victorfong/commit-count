package main

import (
    "testing"
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
