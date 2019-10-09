package aip0203

import (
	"regexp"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/api-linter/lint"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/genproto/googleapis/api/annotations"
)

// This rule inspects the leading comments of each field
// and if anything looks similar to "Immutable.", it throws
// a problem.
//
// Examples:
// Incorrect code for this rule:
//
//	message Book {
//	// The title of the book.
//	// @Immutable
//	string title = 1;
//	}
//
// or
//
//	message Book {
//	// Immutable. The title of the book.
//	string title = 1;
//	}
//
//
// Correct code for this rule:
//
//	message Book {
//	// The title of the book.
//	string title = 1 [(google.api.field_behavior) = IMMUTABLE];
//	}
var immutable = &lint.FieldRule{
	Name:   lint.NewRuleName("core", "0203", "immutable"),
	URI:    "http://api.dev/203#guidance",
	OnlyIf: withoutImmutableFieldBehavior,
	LintField: func(f *desc.FieldDescriptor) []lint.Problem {
		return checkLeadingComments(f, immutableRegexp, "IMMUTABLE")
	},
}

var immutableRegexp = regexp.MustCompile("(?i).*immutable.*")

func withoutImmutableFieldBehavior(f *desc.FieldDescriptor) bool {
	opts := f.GetFieldOptions()
	if fb, err := proto.GetExtension(opts, annotations.E_FieldBehavior); err == nil {
		for _, v := range fb.([]annotations.FieldBehavior) {
			if v == annotations.FieldBehavior_IMMUTABLE {
				return false
			}
		}
	}
	return true
}
