// Copyright 2023 Google LLC
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

package aip0134

import (
	"testing"

	"github.com/googleapis/api-linter/rules/internal/testutils"
	"github.com/jhump/protoreflect/desc"
)

func TestRequiredFieldTests(t *testing.T) {
	for _, test := range []struct {
		name                 string
		Fields               string
		problematicFieldName string
		problems             testutils.Problems
	}{
		{
			"ValidNoExtraFields",
			"",
			"",
			nil,
		},
		{
			"ValidOptionalUpdateMask",
			"google.protobuf.FieldMask update_mask = 2 [(google.api.field_behavior) = OPTIONAL];",
			"update_mask",
			nil,
		},
		{
			"InvalidRequiredValidateOnly",
			"google.protobuf.FieldMask update_mask = 2 [(google.api.field_behavior) = REQUIRED];",
			"update_mask",
			testutils.Problems{
				{Message: `Update RPCs must only require fields explicitly described in AIPs, not "update_mask"`},
			},
		},
		{
			"InvalidRequiredUnknownField",
			"bool create_iam = 3 [(google.api.field_behavior) = REQUIRED];",
			"create_iam",
			testutils.Problems{
				{Message: `Update RPCs must only require fields explicitly described in AIPs, not "create_iam"`},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := testutils.ParseProto3Tmpl(t, `
				import "google/api/annotations.proto";
				import "google/api/field_behavior.proto";
				import "google/api/resource.proto";
				import "google/protobuf/field_mask.proto";

				service Library {
					rpc UpdateBook(UpdateBookRequest) returns (Book) {
						option (google.api.http) = {
							patch: "/v1/{name=publishers/*/books/*}"
							body: "book"
						};
					}
				}

				message Book {
					option (google.api.resource) = {
						type: "library.googleapis.com/Book"
						pattern: "publishers/{publisher}/books/{book}"
					};
					string name = 1;
				}

				message UpdateBookRequest {
					Book book = 1 [
						(google.api.field_behavior) = REQUIRED
					];
					{{.Fields}}
				}
			`, test)
			var dbr desc.Descriptor = f.FindMessage("UpdateBookRequest")
			if test.problematicFieldName != "" {
				dbr = f.FindMessage("UpdateBookRequest").FindFieldByName(test.problematicFieldName)
			}
			if diff := test.problems.SetDescriptor(dbr).Diff(requestRequiredFields.Lint(f)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
