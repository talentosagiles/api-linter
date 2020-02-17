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

package aip0158

import (
	"fmt"

	"github.com/gertd/go-pluralize"
	"github.com/googleapis/api-linter/lint"
	"github.com/googleapis/api-linter/locations"
	"github.com/jhump/protoreflect/desc"
)

var responseFirstFieldMustBePlural = &lint.MessageRule{
	Name:   lint.NewRuleName(158, "response-first-field-must-be-plural"),
	OnlyIf: isPaginatedResponseMessage,
	LintMessage: func(m *desc.MessageDescriptor) []lint.Problem {
		// Rule check: for all messages that end in Response and contain a next_page_token field.
		// Throw a linter warning if, the first field in the message is not named according to plural(message_name.to_snake().split('_')[1:-1]).
		nextPageToken := m.FindFieldByName("next_page_token")
		if nextPageToken == nil {
			return nil
		}

		firstField := m.FindFieldByNumber(1)
		pluralize := pluralize.NewClient()
		if !pluralize.IsPlural(firstField.GetName()) {
			want := pluralize.Plural(firstField.GetName())

			return []lint.Problem{{
				Message:    fmt.Sprintf("The first field should be plural."),
				Suggestion: want,
				Descriptor: firstField,
				Location:   locations.DescriptorName(m),
			}}
		}

		return nil
	},
}