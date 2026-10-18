// Package utils holds console formatting helpers shared by all scenarios.
package utils

import (
	"fmt"
	"os"
	"strings"
)

var useColor = func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}()

const (
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

func paint(color, s string) string {
	if !useColor {
		return s
	}
	return color + s + reset
}

func PrintHeader(title string) { fmt.Println(paint(bold, "=== "+title+" ===")); fmt.Println() }
func PrintSuccess(msg string)  { fmt.Println(paint(green, "✓ "+msg)) }
func PrintError(msg string)    { fmt.Println(paint(red, "✗ "+msg)) }
func PrintWarning(msg string)  { fmt.Println(paint(yellow, "⚠ "+msg)) }
func PrintCritical(msg string) { fmt.Println(paint(bold+red, "❌ "+msg)) }
func PrintSeparator()          { fmt.Println(strings.Repeat("-", 50)) }
