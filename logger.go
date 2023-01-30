package lumina

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

func newTaggedLogger(tags ...string) *log.Logger {
	prefix := fmt.Sprintf("[%08x]", rand.Uint32())
	for _, tag := range tags {
		prefix += fmt.Sprintf("[%s]", tag)
	}
	prefix += " "
	return log.New(os.Stderr, prefix, log.LstdFlags|log.Lmsgprefix)
}

func addLoggerTag(logger *log.Logger, newTags ...string) *log.Logger {
	prefix := ""
	if logger != nil {
		prefix = strings.TrimSuffix(logger.Prefix(), " ")
	}
	for _, tag := range newTags {
		prefix += fmt.Sprintf("[%s]", tag)
	}
	prefix += " "

	var writer io.Writer = os.Stderr
	if logger != nil {
		writer = log.Writer()
	}
	return log.New(writer, prefix, log.LstdFlags|log.Lmsgprefix)
}
