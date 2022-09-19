package lumina

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

func newTaggedLogger(tags ...string) *log.Logger {
	prefix := fmt.Sprintf("[%08x]", rand.Uint32())
	for _, tag := range tags {
		prefix += fmt.Sprintf("[%s]", tag)
	}
	prefix += " "
	return log.New(os.Stderr, prefix, log.LstdFlags)
}
