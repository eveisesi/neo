package tools

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
)

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SlotForFlagID(id uint64) string {

	for slot, flags := range neo.SLOT_TO_FLAGIDS {
		for flag := range flags {
			if flag == id {
				return slot
			}
		}
	}

	return ""

}

func IsGroupAllowed(id uint64) bool {
	for _, v := range neo.ALLOWED_SHIP_GROUPS {
		if v == id {
			return true
		}
	}

	return false
}

func AbbreviateNumber(v float64) string {
	suffix := []string{"", "K", "M", "B", "T"}
	pos := 0

	for v > 999 {
		pos++
		v = v / 1000
	}

	return fmt.Sprintf("%.2f%s", v, suffix[pos])
}

func ToFixed(num float64, precision int) float64 {
	b, e := strconv.ParseFloat(fmt.Sprintf(fmt.Sprintf("%%.%df", precision), num), 64)
	if e != nil {
		return 0.00
	}

	return b
}
