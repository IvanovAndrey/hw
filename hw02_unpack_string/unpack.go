package hw02unpackstring

import (
	"errors"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	if str == "" {
		return "", nil
	}

	runes := []rune(str)
	if !isValidRuneArray(runes) {
		return "", ErrInvalidString
	}

	res := make([]rune, 0, len(runes))
	for i, v := range runes {
		if unicode.IsDigit(v) {
			res = multiply(res, v, runes[i-1])
		} else {
			res = append(res, v)
		}
	}

	return string(res), nil
}

func multiply(res []rune, num rune, symbol rune) []rune {
	number := int(num - '0')
	switch number {
	case 0:
		return res[:len(res)-1]
	default:
		for i := 0; i < number-1; i++ {
			res = append(res, symbol)
		}
	}
	return res
}

func isValidRuneArray(runes []rune) bool {
	if len(runes) > 0 && unicode.IsDigit(runes[0]) {
		return false
	}

	for i := 1; i < len(runes); i++ {
		if unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i-1]) {
			return false
		}
	}

	return true
}
