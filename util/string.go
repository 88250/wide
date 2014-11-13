package util

type str struct{}

// String utilities.
var Str = str{}

// Contains determines whether the str is in the strs.
func (*str) Contains(str string, strs []string) bool {
	for _, v := range strs {
		if v == str {
			return true
		}
	}

	return false
}

// LCS gets the longest common substring of s1 and s2.
//
// Refers to http://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Longest_common_substring.
func (*str) LCS(s1 string, s2 string) string {
	var m = make([][]int, 1+len(s1))

	for i := 0; i < len(m); i++ {
		m[i] = make([]int, 1+len(s2))
	}

	longest := 0
	x_longest := 0

	for x := 1; x < 1+len(s1); x++ {
		for y := 1; y < 1+len(s2); y++ {
			if s1[x-1] == s2[y-1] {
				m[x][y] = m[x-1][y-1] + 1
				if m[x][y] > longest {
					longest = m[x][y]
					x_longest = x
				}
			} else {
				m[x][y] = 0
			}
		}
	}

	return s1[x_longest-longest : x_longest]
}
