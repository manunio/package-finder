package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {

	// set test environment
	if err := os.Setenv("ENV", "test"); err != nil {
		t.Error(err)
	}

	// get []byte for test_package.yml
	// f, err := ioutil.ReadFile("test_package.yml")
	// if err != nil {
	// 	t.Error(err)
	// }

	testPackage := &Package{
		Domains:        []Domain{{Name: "github", UrlsFile: "./urls/github-urls.txt", Enable: true}},
		OutputRootPath: "out",
		InfoLog:        "test_info.log",
		PackageLog:     "test_package.log",
		PackageYml:     "test_package.yml",
	}

	tests := map[string]struct {
		input  *Package
		output *Package
		fails  bool
	}{
		"fails on missing package.yml path": {
			input:  &Package{PackageYml: ""},
			output: &Package{PackageYml: ""},
			fails:  true,
		},
		"passes on package.yml path & marshal": {
			input:  &Package{PackageYml: "test_package.yml"},
			output: testPackage,
			fails:  false,
		},
	}
	for _, test := range tests {
		err := test.input.readConfig()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, test.input)
	}
}

// func TestValidateConfig(t * testing.T) error {
//
// }
