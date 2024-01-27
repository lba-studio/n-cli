package utils

import (
	"errors"
	"strings"
)

const MessagePlaceholder = "{{message}}"

var (
	ErrMessageFormatMissingPlaceholder = errors.New("{{message}} placeholder is missing from messageFormat")
)

func GetMessageFromFormat(format, msg string) (string, error) {
	if format == "" {
		return msg, nil
	}
	if !strings.Contains(format, MessagePlaceholder) {
		return "", ErrMessageFormatMissingPlaceholder
	}
	msg = strings.Replace(format, MessagePlaceholder, msg, 1)
	return msg, nil
}
