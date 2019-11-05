// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each case must be positive when the rule in test
// is enabled. It must also contain a "disable-me-here"
// comment at the place where you want the rule to be
// disabled.
var testCases = []struct {
	testName, rule, proto string
}{
	{
		testName: "GetRequestMessage",
		rule:     "core::0131::request-message-name",
		proto: `
		syntax = "proto3";

		service Library {
			// disable-me-here
			rpc GetBook(Book) returns (Book);
		}

		message Book {}
		`,
	},
	{
		testName: "FieldNames",
		rule:     "core::0140::lower-snake",
		proto: `
				syntax = "proto3";
				import "dummy.proto";
				message Test {
					// disable-me-here
					string badName = 1;
					dummy.Dummy dummy = 2;
				}
			`,
	},
}

func TestRules_EnabledByDefault(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.testName, func(t *testing.T) {
			proto := test.proto
			result := runLinter(t, proto, "")
			if !strings.Contains(result, test.rule) {
				t.Errorf("Rule %q should be enabled by default", test.rule)
			}
		})
	}
}

func TestRules_DisabledByFileComments(t *testing.T) {
	config := `
	[
		{
			"included_paths": ["*.proto"],
			"rule_configs": {
				"": {
					"status": "enabled",
					"category": "warning"
				}
			}
		}
	]
	`

	for _, test := range testCases {
		t.Run(test.testName, func(t *testing.T) {
			disableInFile := fmt.Sprintf("// (-- api-linter: %s=disabled --)", test.rule)
			proto := disableInFile + "\n" + test.proto
			result := runLinter(t, proto, config)
			if strings.Contains(result, test.rule) {
				t.Errorf("Rule %q should be disabled by file comments", test.rule)
			}
		})
	}
}

func TestRules_DisabledByInlineComments(t *testing.T) {
	config := `
	[
		{
			"included_paths": ["*.proto"],
			"rule_configs": {
				"": {
					"status": "enabled",
					"category": "warning"
				}
			}
		}
	]
	`

	for _, test := range testCases {
		t.Run(test.testName, func(t *testing.T) {
			disableInline := fmt.Sprintf("(-- api-linter: %s=disabled --)", test.rule)
			proto := strings.Replace(test.proto, "disable-me-here", disableInline, -1)
			result := runLinter(t, proto, config)
			if strings.Contains(result, test.rule) {
				t.Errorf("Rule %q should be disabled by in-line comments", test.rule)
			}
		})
	}
}

func TestRules_DisabledByConfig(t *testing.T) {
	config := `
	[
		{
			"included_paths": ["*.proto"],
			"rule_configs": {
				"": {
					"disabled": false,
					"category": "warning"
				}
			}
		},
		{
			"included_paths": ["*.proto"],
			"rule_configs": {
				"replace-me-here": {
					"disabled": true
				}
			}
		}
	]
	`

	for _, test := range testCases {
		t.Run(test.testName, func(t *testing.T) {
			c := strings.Replace(config, "replace-me-here", test.rule, -1)
			result := runLinter(t, test.proto, c)
			if strings.Contains(result, test.rule) {
				t.Errorf("Rule %q should be disabled by the user config: %q", test.rule, c)
			}
		})
	}
}

func runLinter(t *testing.T, protoContent, configContent string) string {
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Prepare command line flags.
	args := []string{}
	// Add a flag for the linter config file if the provided
	// config content is not empty.
	if configContent != "" {
		configFileName := "test_config.json"
		configFilePath := filepath.Join(tempDir, configFileName)
		if err := writeFile(configFilePath, configContent); err != nil {
			t.Fatal(err)
		}
		args = append(args, fmt.Sprintf("--config=%s", configFilePath))
	}
	// Add a flag for the output path.
	outPath := filepath.Join(tempDir, "test.out")
	args = append(args, fmt.Sprintf("-o=%s", outPath))
	// Add the temp dir to the proto paths.
	args = append(args, fmt.Sprintf("-I=%s", tempDir))
	// Add a flag for the file descriptor set.
	args = append(args, "--proto-descriptor-set=internal/testdata/dummy.protoset")
	// Write the proto file.
	protoFileName := "test.proto"
	protoFilePath := filepath.Join(tempDir, protoFileName)
	if err := writeFile(protoFilePath, protoContent); err != nil {
		t.Fatal(err)
	}
	args = append(args, protoFileName)

	if err := runCLI(args); err != nil {
		t.Fatal(err)
	}

	out, err := ioutil.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func writeFile(path, content string) error {
	if path == "" {
		return nil
	}
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(content), 0644)
}
