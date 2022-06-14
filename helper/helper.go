// Useful helper functions
package helper

import "math/rand"

// Generate a random string  of "n" length, composed of uppercase letters and numbers
func RandomString(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
