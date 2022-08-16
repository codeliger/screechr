package main

import "strconv"

// parseCount parses and returns the amount of screeches to display per page
func parseCount(countRaw string) int {
	count, err := strconv.Atoi(countRaw)
	if err != nil || count < 1 {
		count = 50
	}

	if count > 500 {
		count = 500
	}

	return count
}

// anyEmpty returns true if any of the given strings are empty
func anyEmpty(strs ...string) bool {
	for _, str := range strs {
		if str == "" {
			return true
		}
	}

	return false
}
