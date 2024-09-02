package score

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/h2non/filetype"
)

type Options struct {
	OS         []string
	Arch       []string
	Extensions []string
	Names      []string
}

func Score(names []string, opts *Options) []Sorted {
	var scores = make(map[string]int)

	for _, name := range names {
		var score int
		var scoringValues = make(map[string]int)

		// Note: if it has the word "update" in it, we want to deprioritize it as it's likely an update binary from
		// a rust or go binary distribution
		scoringValues["update"] = -100

		for _, os1 := range opts.OS {
			scoringValues[strings.ToLower(os1)] = 40
		}
		for _, arch := range opts.Arch {
			scoringValues[strings.ToLower(arch)] = 30
		}
		for _, ext := range opts.Extensions {
			scoringValues[strings.ToLower(ext)] = 20
		}
		for _, name1 := range opts.Names {
			scoringValues[strings.ToLower(name1)] = 10
		}

		for keyMatch, keyScore := range scoringValues {
			if keyScore == 20 { // handle extensions special
				if ext := strings.TrimPrefix(filepath.Ext(strings.ToLower(name)), "."); ext != "" {
					for _, fileExt := range opts.Extensions {
						if filetype.GetType(ext) == filetype.GetType(fileExt) {
							score += keyScore
							break
						}
					}
				}
			} else {
				if strings.Contains(strings.ToLower(name), keyMatch) {
					score += keyScore
				}
			}
		}

		scores[name] = score
	}

	return sortMapByValue(scores)
}

type Sorted struct {
	Key   string
	Value int
}

func sortMapByValue(m map[string]int) []Sorted {
	var sorted []Sorted

	// Create a slice of key-value pairs
	for k, v := range m {
		sorted = append(sorted, struct {
			Key   string
			Value int
		}{k, v})
	}

	// Sort the slice based on the values in descending order
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	return sorted
}
