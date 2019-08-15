// Package parser is a file name parser
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var matcher *regexp.Regexp

var (
	romanNumerals = map[rune]int{
		'M': 1000,
		'D': 500,
		'C': 100,
		'L': 50,
		'X': 10,
		'V': 5,
		'I': 1,
	}

	notAllowedSubStrings = []string{
		"NCED",
		"NCOP",
		"Commic",
		"OVA",
	}
)

// romanToInt converts a roman numeral to an integer
func romanToInt(numerals string) (int, error) {
	// use runes in case we get non ASCII input
	runes := []rune(strings.ToUpper(numerals))
	strLen := len(runes)

	num := 0
	lastValue := 0
	for i := 0; i != strLen; i++ {
		// get the rune
		r := rune(numerals[i])

		v := romanNumerals[r]
		if v == 0 { // skip unknowns
			return 0, fmt.Errorf("invalid roman numeral %c", r)
		}

		// if the last value was less, then we should subtract it
		if lastValue < v {
			// we added it last iteration, so subtract it twice
			num -= lastValue * 2
		}

		num += v

		fmt.Printf("num is %d\n", num)

		lastValue = v
	}

	return num, nil
}

// Metadata is information returned by the parser
type Metadata struct {
	// Season is the season of this media
	Season int

	// Episode is the episode of this media
	Episode int
}

// ParseFile returns detected metadata from a file name
func ParseFile(name string) (Metadata, error) {
	matches := matcher.FindStringSubmatch(name)
	if len(matches) == 0 {
		return Metadata{}, fmt.Errorf("failed to find a match")
	}

	season := 1
	if matches[1] != "" {
		var err error
		season, err = strconv.Atoi(matches[1])
		if err != nil { // try to parse roman numeral
			season, err = romanToInt(matches[1])
			if err != nil {
				return Metadata{}, err
			}
		}
	}

	episode, err := strconv.Atoi(matches[2])
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to parse episode number: %v", err)
	}

	for _, substr := range notAllowedSubStrings {
		i := strings.Index(name, substr)
		if i != -1 {
			return Metadata{}, fmt.Errorf("found not allowed substr in title '%s'", substr)
		}
	}

	return Metadata{
		Season:  season,
		Episode: episode,
	}, nil
}

func init() {
	matcher = regexp.MustCompile(`(\d+|[iI]+)?(?: -)?(?:[eE _x[]|^)(\d+)[^x]`)
}
