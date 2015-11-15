package checks

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

func hashFromf(s string, p ...interface{}) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf(s, p...)))
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return hash
}
