package restore_ip_addresses

import (
	"strconv"
	"strings"
)

func restoreIpAddresses(s string) []string {
	res := make([]string, 0)
	// pos is going to be the position in 's'
	// each element in path is a generated segment
	var backtrack func(pos int, path []string)
	backtrack = func(pos int, path []string) {
		if len(path) == 4 { // we've generated 4 segments
			if pos == len(s) {
				validPath := strings.Join(path, ".")
				res = append(res, validPath)
			}
			return
		}

		for end := pos; end < len(s) && end < pos+3; end++ {
			segment := s[pos : end+1]
			if !isValid(segment) {
				continue
			}

			path = append(path, segment)
			backtrack(end+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]string, 0))

	return res
}

func isValid(segment string) bool {
	if len(segment) == 0 {
		return false
	}

	if segment[0] == '0' {
		if len(segment) > 1 {
			return false
		} else { // length 1 segment with leading 0 is valid
			return true
		}
	}

	octet, _ := strconv.Atoi(segment)
	if octet > 255 {
		return false
	}

	return true
}
