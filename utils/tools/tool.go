package tools

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strings"
	"time"
)

func UUIDHex() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func GetAllLike(kw string) string {
	return "%" + kw + "%"
}

func GetPrefixLike(kw string) string {
	return "%" + kw
}

func GetSuffixLike(kw string) string {
	return kw + "%"
}
