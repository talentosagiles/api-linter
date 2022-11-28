// Copyright 2022 Google LLC
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

package aip0127

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/googleapis/api-linter/lint"
	"github.com/googleapis/api-linter/rules/internal/utils"
	"github.com/jhump/protoreflect/desc"
)

var (
	resourcePatternSegment     = `[^/]+`
	resourcePatternAnySegments = fmt.Sprintf("((%s/)*%s)?", resourcePatternSegment, resourcePatternSegment)
	pathTemplateToRegex        = strings.NewReplacer("**", resourcePatternAnySegments, "*", resourcePatternSegment)
)

type resourceReference struct {
	// The path of the field with the `google.api.resource_reference`. This is
	// provided as a variable in the HTTPRule.
	fieldPath string
	// A template that the resource's pattern string must adhere to. This is
	// provided by the variable's template in the HTTPRule.
	pathTemplate string
	// The name of the resource type. This is used to look up the resource
	// message.
	resourceRefName string
}

// Returns a list of resourceReferences for each variable in all the method's
// HTTPRule's.
func resourceRefsForMethod(m *desc.MethodDescriptor) []resourceReference {
	resourceRefs := []resourceReference{}
	for _, httpRule := range utils.GetHTTPRules(m) {
		resourceRefs = append(resourceRefs, resourceRefsForHttpRule(httpRule, m.GetInputType())...)
	}
	return resourceRefs
}

// Returns a resourceReference for every variable in the given HTTPRule.
func resourceRefsForHttpRule(httpRule *utils.HTTPRule, msg *desc.MessageDescriptor) []resourceReference {
	resourceRefs := []resourceReference{}
	for fieldPath, template := range httpRule.GetVariables() {
		// Find the (sub-)field in the message corresponding to the variable's
		// field path.
		if resourceRefField := utils.FindFieldDotNotation(msg, fieldPath); resourceRefField != nil {
			// Extract the name of the resource referenced by this field.
			if resourceRef := utils.GetResourceReference(resourceRefField); resourceRef != nil {
				// TODO(acamadeo): Support the case where a resource has
				// multiple parent resources.
				if resourceRef.GetChildType() != "" {
					continue
				}

				res := resourceReference{fieldPath: fieldPath, pathTemplate: template, resourceRefName: resourceRef.GetType()}
				resourceRefs = append(resourceRefs, res)
			}
		}
	}
	return resourceRefs
}

// Constructs a regex from the HTTPRule's path template representing resource
// patterns that it will match against.
func compilePathTemplateRegex(pathTemplate string) (*regexp.Regexp, error) {
	pattern := fmt.Sprintf("^%s$", pathTemplateToRegex.Replace(pathTemplate))
	return regexp.Compile(pattern)
}

func anyStringsMatchRegex(regex *regexp.Regexp, strs []string) bool {
	for _, str := range strs {
		if regex.MatchString(str) {
			return true
		}
	}
	return false
}

// Checks whether the HTTP pattern specified in `resourceRef` matches any of the
// patterns defined for that resource.
func checkHttpPatternMatchesResource(m *desc.MethodDescriptor, resourceRef resourceReference) []lint.Problem {
	annotation := utils.FindResource(resourceRef.resourceRefName, m.GetFile())
	if annotation == nil {
		message := fmt.Sprintf("Unable to find resource with name %q", resourceRef.resourceRefName)
		return []lint.Problem{{Message: message, Descriptor: m.GetInputType()}}
	}

	pathRegex, err := compilePathTemplateRegex(resourceRef.pathTemplate)
	if err != nil {
		message := fmt.Sprintf("HTTP annotation includes an invalid path template %q", resourceRef.pathTemplate)
		return []lint.Problem{{Message: message, Descriptor: m}}
	}

	if !anyStringsMatchRegex(pathRegex, annotation.GetPattern()) {
		message := fmt.Sprintf("The HTTP pattern %q does not match any of the patterns for resource %q", resourceRef.pathTemplate, resourceRef.resourceRefName)
		return []lint.Problem{{Message: message, Descriptor: m}}
	}

	return []lint.Problem{}
}

var httpTemplatePattern = &lint.MethodRule{
	Name: lint.NewRuleName(127, "http-template-pattern"),
	OnlyIf: func(m *desc.MethodDescriptor) bool {
		return len(resourceRefsForMethod(m)) > 0
	},
	LintMethod: func(m *desc.MethodDescriptor) []lint.Problem {
		problems := []lint.Problem{}

		resourceRefs := resourceRefsForMethod(m)
		for _, resourceRef := range resourceRefs {
			problems = append(problems, checkHttpPatternMatchesResource(m, resourceRef)...)
		}

		return problems
	},
}
