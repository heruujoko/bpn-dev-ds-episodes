// Package utils holds console formatting helpers shared by all scenarios.
package utils

import (
	"fmt"
	"os"
	"strings"
)

// useColor: only emit ANSI codes when stdout is a real terminal and NO_COLOR is unset,
// so piped output (and test logs) stay clean.
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

// PrintHeader prints a section title between "===" markers.
func PrintHeader(title string) {
	fmt.Println(paint(bold, "=== "+title+" ==="))
	fmt.Println()
}

// PrintSuccess: ✓ (e.g. cache hit, good outcome).
func PrintSuccess(msg string) { fmt.Println(paint(green, "✓ "+msg)) }

// PrintError: ✗ (e.g. cache miss).
func PrintError(msg string) { fmt.Println(paint(red, "✗ "+msg)) }

// PrintWarning: ⚠ (e.g. stale data).
func PrintWarning(msg string) { fmt.Println(paint(yellow, "⚠ "+msg)) }

// PrintCritical: ❌ (a bug caused by caching).
func PrintCritical(msg string) { fmt.Println(paint(bold+red, "❌ "+msg)) }

// PrintStats prints hit/miss counts and the hit ratio.
func PrintStats(hits, misses int64) {
	ratio := 0.0
	if total := hits + misses; total > 0 {
		ratio = float64(hits) / float64(total) * 100
	}
	fmt.Printf("Stats: %d hits, %d misses, %.0f%% hit ratio\n", hits, misses, ratio)
}

// PrintSeparator prints a visual divider.
func PrintSeparator() {
	fmt.Println(strings.Repeat("-", 50))
}
