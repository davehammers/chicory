// util package contains common/convenient routines to aid in development and debugging
package util

import (
	"fmt"
	"strings"
	"time"
)

// Convert an []int to a single string
func IntArrayToString(a []int64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

// Boolp converts the bool to *bool
func Boolp(in bool) *bool {
	return &in
}

// Stringp converts the string to *string
func Stringp(in string) *string {
	return &in
}

// Intp converts the int to *int
func Intp(in int) *int {
	return &in
}

// Int64p converts the int64 to *int64
func Int64p(in int64) *int64 {
	return &in
}

// Timep converts the time.Time to *time.Time
func Timep(in time.Time) *time.Time {
	return &in
}
