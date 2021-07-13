package commands

import (
	"strings"
)

func WordWrap(str string, size int) (retval []string) {
	if len(str) <= size {
		return []string{str}
	}
	splits := strings.Split(str, " ")
	acc := ""
	for _, split := range splits {
		if len(acc+split) > size {
			if trim := strings.TrimSpace(acc); len(trim) > 0 {
				retval = append(retval, trim[:])
			}
			acc = split
		} else {
			acc += " " + split
		}
	}
	if trim := strings.TrimSpace(acc); len(trim) > 0 {
		retval = append(retval, trim[:])
	}
	return
}

func In(elem string, slice []string) bool {
	for _, e := range slice {
		if elem == e {
			return true
		}
	}
	return false
}
