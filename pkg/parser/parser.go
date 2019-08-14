// Package parser is a file name parser
package parser

import (
	"regexp"
)

var matcher *regexp.Regexp

type Metadata struct {
	// Season is the season of this media
	Season int

	// Episode is the episode of this media
	Episode int
}

func ParseFile(name string) Metadata {
	matcher.FindAllString(name)
}

func init() {
	matcher = regexp.MustCompile(`/(\d+|i+)?(?: -)?(?:[e _x[]|^)(\d+)(?!x)/`)
}
