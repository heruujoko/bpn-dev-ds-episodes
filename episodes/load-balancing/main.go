package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"load-balancing-demo/scenarios"
)

func main() {
	list := []struct {
		name string
		run  func() error
	}{
		{"One server -> round-robin", scenarios.RoundRobin},
		{"Different capacities -> weighted routing", scenarios.Weighted},
	}
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== System Design Episodes: Load Balancing ===")
		for i, s := range list {
			fmt.Printf("%d. %s\n", i+1, s.name)
		}
		fmt.Println("0. Exit")
		fmt.Printf("Choose scenario (0-%d): ", len(list))
		line, err := in.ReadString('\n')
		choice, convErr := strconv.Atoi(strings.TrimSpace(line))
		switch {
		case convErr == nil && choice == 0:
			return
		case convErr == nil && choice >= 1 && choice <= len(list):
			fmt.Println()
			if runErr := list[choice-1].run(); runErr != nil {
				fmt.Println("error:", runErr)
			}
			fmt.Print("\nPress Enter to return to the menu...")
			if _, pauseErr := in.ReadString('\n'); pauseErr != nil {
				fmt.Println()
				return
			}
		case err != nil:
			fmt.Println()
			return
		default:
			fmt.Printf("Invalid choice %q, enter 0-%d.\n", strings.TrimSpace(line), len(list))
		}
	}
}
