package main

import (
	"bufio"
	"circuit-breaker-demo/scenarios"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func menu(input io.Reader, output io.Writer) {
	list := []struct {
		name string
		run  func(io.Writer) error
	}{
		{"Healthy checkout: three couriers", scenarios.Healthy},
		{"B down: no breaker", scenarios.NoBreaker},
		{"Wrong boundary: shared breaker", scenarios.Shared},
		{"Per-courier isolation and recovery", scenarios.Isolated},
	}
	in := bufio.NewReader(input)
	for {
		fmt.Fprintln(output, "\n=== Circuit Breaker: Three Courier Partners ===")
		for i, s := range list {
			fmt.Fprintf(output, "%d. %s\n", i+1, s.name)
		}
		fmt.Fprint(output, "0. Exit\nChoose scenario (0-4): ")
		line, err := in.ReadString('\n')
		n, parseErr := strconv.Atoi(strings.TrimSpace(line))
		switch {
		case parseErr == nil && n == 0:
			return
		case parseErr == nil && n >= 1 && n <= len(list):
			fmt.Fprintln(output)
			if err := list[n-1].run(output); err != nil {
				fmt.Fprintln(output, "error:", err)
			}
			fmt.Fprint(output, "\nPress Enter to return to the menu...")
			if _, err := in.ReadString('\n'); err != nil {
				fmt.Fprintln(output)
				return
			}
		case err != nil:
			fmt.Fprintln(output)
			return
		default:
			fmt.Fprintf(output, "Invalid choice %q; enter 0-4.\n", strings.TrimSpace(line))
		}
	}
}
func main() { menu(os.Stdin, os.Stdout) }
