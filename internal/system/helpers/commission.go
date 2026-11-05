package helpers

import "strconv"

func FormatCommissionRate(basisPoints int) string {
	return strconv.FormatFloat(float64(basisPoints)/100, 'f', -1, 64)
}
