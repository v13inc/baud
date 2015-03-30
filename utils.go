package main

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	NBSP = '\u00A0'
)

func runes(str string) []rune {
	return bytes.Runes([]byte(str))
}

func count(str string) int {
	return utf8.RuneCount([]byte(str))
}

func truncate(str string, num int) string {
	runes := runes(str)
	if len(runes) > num {
		return string(runes[:num])
	}

	return str
}

func pad(str string, char rune, count int) string {
	runes := runes(str)
	diff := count - len(runes)
	if count <= 0 {
		return truncate(str, len(runes))
	}

	return str + strings.Repeat(string(char), diff)
}
