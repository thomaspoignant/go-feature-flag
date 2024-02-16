package ofrep

import (
	"encoding/json"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"hash/crc32"
)

func flagCheckSum(f flag.Flag) string {
	jsonData, _ := json.Marshal(f)
	return fmt.Sprintf("%x", crc32.ChecksumIEEE(jsonData))
}
