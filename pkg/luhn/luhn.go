package luhn

import "strings"

// Valid checks if a card number is valid
func Valid(numberString string) bool {
	card := strings.ReplaceAll(numberString, " ", "")
	sum := 0
	nDigits := len(card)
	parity := nDigits % 2
	for i := 0; i <= nDigits-1; i++ {
		digit := int(card[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
