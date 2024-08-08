// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aip0235

import (
	"testing"

	"github.com/googleapis/api-linter/rules/internal/testutils"
)

func TestRequestUnknownFields(t *testing.T) {
	for _, test := range []struct {
		name        string
		MessageName string
		FieldName   string
		problems    testutils.Problems
	}{
		{"Valid-AllowMissing", "BatchDeleteBooks", "allow_missing", testutils.Problems{}},
		{"Valid-Force", "BatchDeleteBooks", "force", testutils.Problems{}},
		{"Valid-Names", "BatchDeleteBooks", "names", testutils.Problems{}},
		{"Valid-Parent", "BatchDeleteBooks", "parent", testutils.Problems{}},
		{"Valid-RequestID", "BatchDeleteBooks", "request_id", testutils.Problems{}},
		{"Valid-Requests", "BatchDeleteBooks", "requests", testutils.Problems{}},
		{"Valid-ValidateOnly", "BatchDeleteBooks", "validate_only", testutils.Problems{}},
		{"Invalid", "BatchDeleteBooks", "foo", testutils.Problems{{Message: "Unexpected field"}}},
		{"IrrelevantMessage", "DeleteBooks", "foo", testutils.Problems{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := testutils.ParseProto3Tmpl(t, `
				message {{.MessageName}}Request {
					repeated string {{.FieldName}} = 1;
				}
			`, test)
			field := f.GetMessageTypes()[0].GetFields()[0]
			if diff := test.problems.SetDescriptor(field).Diff(requestUnknownFields.Lint(f)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
