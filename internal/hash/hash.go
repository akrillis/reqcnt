package hash

import (
	cSHA512 "crypto/sha512"
	"encoding/hex"
)

func Hash(data string) string {
	hh := cSHA512.New()
	hh.Write([]byte(data))
	return hex.EncodeToString(hh.Sum(nil))
}
