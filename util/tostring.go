package util

import "strconv"

// PriceToString converts price float64) to string
func PriceToString(n float64) string  {
  return strconv.FormatFloat(n, 'f', 3, 64)
}

// AmountToString converts price float64) to string
func AmountToString(n float64) string  {
  return strconv.FormatFloat(n, 'f', 8, 64)
}