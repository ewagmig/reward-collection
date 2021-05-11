package version

import "fmt"

// Variables defined by the Makefile and passed in with ldflags
var ProgramName string
var Main string
var ChangeLog string
var BuiltAt string

// Version returns the application server version
func Version() string {
	return fmt.Sprintf("Main Version: %s\nChangeLog: %s\nBuilt At: %s", Main, ChangeLog, BuiltAt)
}
