package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseFileTestStruct struct {
	name string
	Metadata
}

func TestRomanNumerals(t *testing.T) {
	tests := map[string]int{
		"V":   5,
		"II":  2,
		"XIV": 14,
		"XVI": 16,
	}

	for numeral, expectedVal := range tests {
		v, err := romanToInt(numeral)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, v, fmt.Sprintf("roman numeral %s got wrong value", numeral))
	}

	return
}

func TestParseFile(t *testing.T) {
	tests := []parseFileTestStruct{
		parseFileTestStruct{
			name: "Ansatsu Kyoushitsu 2x17.mkv",
			Metadata: Metadata{
				Episode: 17,
				Season:  2,
			},
		},
		parseFileTestStruct{
			name: "[Prefix] Hello World 1x12 (1920x1080).mkv",
			Metadata: Metadata{
				Episode: 12,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "[Lit Club] ASsdfasd x12 (193232x13194).mkv",
			Metadata: Metadata{
				Episode: 12,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "[Prefix]Gintama_-_004_(10bit_BD720p_x265).mkv",
			Metadata: Metadata{
				Episode: 4,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "Welcome_to_the_NHK_01_(DVD_480p)_(Suffix).mkv",
			Metadata: Metadata{
				Episode: 1,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "[Prefix] KonoSuba II 10.mkv",
			Metadata: Metadata{
				Episode: 10,
				Season:  2,
			},
		},
		parseFileTestStruct{
			name: "[Prefix] Sword Art Online II - 01v2mkv",
			Metadata: Metadata{
				Episode: 1,
				Season:  2,
			},
		},
		parseFileTestStruct{
			name: "100. Hello [BD 1080p Hi10p AAC dual-audio][world].mkv",
			Metadata: Metadata{
				Episode: 100,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "[HelloWorld][100][Fuck][My][Life][How][Did][This][Seem][Like][A][Good][IDEA].mkv",
			Metadata: Metadata{
				Episode: 100,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "30-my-name-has-a-number-because-im-special 2.mkv",
			Metadata: Metadata{
				Episode: 2,
				Season:  1,
			},
		},
		parseFileTestStruct{
			name: "my-name-30-has-a-number-because-im-special 2.mkv",
			Metadata: Metadata{
				Episode: 2,
				Season:  1,
			},
		},
	}

	for _, test := range tests {
		m, err := ParseFile(test.name)
		assert.NoError(t, err)
		assert.Equal(t, test.Metadata, m, fmt.Sprintf("file name '%s' failed", test.name))
	}
}

func TestParseFileSkip(t *testing.T) {
	tests := []parseFileTestStruct{
		parseFileTestStruct{
			name: "[Prefix] Some OVA (1080p Blu-ray 8bit AAC).mp4",
		},
		parseFileTestStruct{
			name: "[PREFIX] NCED.mkv",
		},
		parseFileTestStruct{
			name: "[PREFIX] NCEDv2.mkv",
		},
		parseFileTestStruct{
			name: "[Prefix] Bad Show - WHAT IS GOING ON? [HASH?].mkv",
		},
		parseFileTestStruct{
			name: "[PREFIX] NCOP.mkv",
		},
	}

	for _, test := range tests {
		m, err := ParseFile(test.name)
		assert.Error(t, err)
		assert.Equal(t, test.Metadata, m, fmt.Sprintf("file name '%s' failed", test.name))
	}
}
