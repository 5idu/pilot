package xstring

import (
	"fmt"

	"github.com/5idu/pilot/pkg/util/xcrypto"
)

// GenerateUUID simply generates an unique UID.
func GenerateUUID() string {
	buf, err := xcrypto.Bytes(16)
	if err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

// Short is used to generate the first 8 characters of a UUID.
func GenerateShortUUID() string {
	return GenerateUUID()[0:8]
}
