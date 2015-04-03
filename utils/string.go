package utils

import (
	"bytes"
	"math"
	"strings"
	"unicode/utf8"
)

const (
	NBSP = '\u00A0'
)

/*
	String Utils
*/
func Runes(str string) []rune {
	return bytes.Runes([]byte(str))
}

func Count(str string) int {
	return utf8.RuneCount([]byte(str))
}

func Truncate(str string, num int) string {
	runes := Runes(str)
	if len(runes) > num {
		return string(runes[:num])
	}

	return str
}

func Pad(str string, char rune, count int) string {
	runes := Runes(str)
	diff := count - len(runes)
	if diff <= 0 {
		return Truncate(str, len(runes))
	}

	return str + strings.Repeat(string(char), diff)
}

func Center(str string, char rune, count int) string {
	c := string(char)
	chars := len(Runes(str))
	if chars >= count {
		return Truncate(str, count)
	} else {
		diff := float64(count - chars)
		p := int(math.Floor(diff / 2))
		r := int(math.Ceil(math.Mod(diff, 2)))
		return strings.Repeat(c, p) + str + strings.Repeat(c, p) + strings.Repeat(c, r)
	}
}
