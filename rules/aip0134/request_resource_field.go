// Copyright 2019 Google LLC
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
	"fmt"

	"github.com/googleapis/api-linter/lint"
	"github.com/jhump/protoreflect/desc"
	"github.com/stoewer/go-strcase"
)

// The create request message should have resource field.
var resourceField = &lint.MessageRule{
	Name:   lint.NewRuleName(134, "request--resource-field"),
	OnlyIf: isUpdateRequestMessage,
	LintMessage: func(m *desc.MessageDescriptor) []lint.Problem {
		resourceMsgName := extractResource(m.GetName())

		// The rule (resource field name must map to the POST body) is
		// checked by AIP-0134 ("core::0134::http-body")
		for _, fieldDesc := range m.GetFields() {
			if msgDesc := fieldDesc.GetMessageType(); msgDesc != nil && msgDesc.GetName() == resourceMsgName {
				// Rule check: Is the field named properly?
				if want := strcase.SnakeCase(resourceMsgName); fieldDesc.GetName() != want {
					return []lint.Problem{{
						Message:    fmt.Sprintf("Resource field should be named %q.", want),
						Descriptor: fieldDesc,
						Suggestion: want,
						Location:   lint.DescriptorNameLocation(fieldDesc),
					}}
				}
				return nil
			}
		}

		// Rule check: Establish that a resource field must be included.
		return []lint.Problem{{
			Message:    fmt.Sprintf("Message %q has no %q type field", m.GetName(), resourceMsgName),
			Descriptor: m,
		}}
	},
}
