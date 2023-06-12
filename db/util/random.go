package util

import (
	"math/rand"
	"strings"
	"time"
)
const alphabet = "abcdefghijklmnopqrstuvwxyz"
// called when random.go is run
func init() {
	rand.Seed(time.Now().UnixNano()) // int64
}
// generate random integer between min and max
func RandInt(min, max int64) int64{
	return min + rand.Int63n(max - min + 1)
	// pseudo-random generator between min->max
}
// generate random string of given length
func RandomString(n int) string{
	// string builder
	var sb strings.Builder
	k := len(alphabet)
	for i:=0; i<n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
// generate random owner name (length of 6 characters)
func RandomOwner() string{
	return RandomString(6)
}
// random balance generator between min and max
func RandomMoney() int64{
	return RandInt(0,1000)
}
// select a random currency out of a list
func RandomCurrency() string {
	currencies:= []string{"EUR", "USD", "CAD"}
	n:= len(currencies)
	return currencies[rand.Intn(n)]
}