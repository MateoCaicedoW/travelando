package helpers

import "strconv"

func FormatCOP(amount int64) string {
	value := strconv.FormatInt(amount, 10)
	start := 0
	if len(value) > 0 && value[0] == '-' {
		start = 1
	}
	for i := len(value) - 3; i > start; i -= 3 {
		value = value[:i] + "." + value[i:]
	}
	return value
}
