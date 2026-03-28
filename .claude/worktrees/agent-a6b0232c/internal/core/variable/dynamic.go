package variable

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// ExpandDynamic replaces dynamic variable placeholders with generated values.
// Only values that exactly match a dynamic key are expanded; partial matches
// are left unchanged.
func ExpandDynamic(vars map[string]string) map[string]string {
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		switch v {
		case "$timestamp":
			result[k] = strconv.FormatInt(time.Now().Unix(), 10)
		case "$isoTimestamp":
			result[k] = time.Now().UTC().Format(time.RFC3339)
		case "$randomInt":
			result[k] = randomInt(0, 1001) // [0, 1000]
		case "$randomUUID":
			result[k] = generateUUIDv4()
		case "$guid":
			result[k] = generateUUIDv4()
		default:
			result[k] = v
		}
	}
	return result
}

// randomInt returns a random integer in [min, max) as a string.
func randomInt(min, max int) string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(n.Int64()+int64(min), 10)
}

// generateUUIDv4 generates a UUID v4 using crypto/rand.
func generateUUIDv4() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant bits
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
