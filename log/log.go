package log

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Prefixes in log.
const (
	PrefixDebug = "[DEBUG] "
	PrefixInfo  = "[INFO] "
	PrefixWarn  = "[WARN] "
	PrefixError = "[ERROR] "
)

var (
	debug = false
	out   = os.Stdout
)

func init() {
	debug = viper.GetBool("debug")
}

// Debugf formats according to a format specifier with PrefixDebug string.
func Debugf(format string, a ...interface{}) {
	if !debug {
		return
	}

	fmt.Fprintf(out, PrefixDebug+format, a...)
}

// Infof formats according to a format specifier with PrefixInfo string.
func Infof(format string, a ...interface{}) {
	fmt.Fprintf(out, PrefixInfo+format, a...)
}

// Warnf formats according to a format specifier with PrefixWarn string.
func Warnf(format string, a ...interface{}) {
	fmt.Fprintf(out, PrefixWarn+format, a...)
}

// Errorf formats according to a format specifier with PrefixError string.
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(out, PrefixError+format, a...)
}
