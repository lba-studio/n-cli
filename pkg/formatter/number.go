package formatter

import (
	"strconv"
)

func PrettyPrintInt64(num int64) string {
	isNegative := false
	if num < 0 {
		isNegative = true
		num = -num
	}
	numberTokens := make([]string, 0)
	for {
		numStr := strconv.FormatInt(num%1000, 10)
		if num < 1000 {
			numberTokens = append(numberTokens, numStr)
			break
		}
		for len(numStr) < 3 {
			numStr = "0" + numStr
		}
		numberTokens = append(numberTokens, numStr)
		num = num / 1000
	}
	// convert to string
	out := ""
	for _, token := range numberTokens {
		if out == "" {
			out = token
		} else {
			out = token + "," + out
		}
	}
	if isNegative {
		return "-" + out
	} else {
		return out
	}
}
