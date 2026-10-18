// Rate limit demo: an interactive menu that runs rate-limiting scenarios.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"rate-limit-demo/scenarios"
)

type Scenario interface {
	Name() string
	Description() string
	Run() error
}

func main() {
	list := []Scenario{scenarios.NoLimitScenario{}, scenarios.FixedWindowScenario{}, scenarios.TokenBucketScenario{}, scenarios.KeyChoiceScenario{}, scenarios.MultiServerScenario{}, scenarios.ScaleOutScenario{}}
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("=== Rate Limit Design System Episodes ===")
		fmt.Println()
		for i, s := range list {
			fmt.Printf("%d. %s\n", i+1, s.Name())
		}
		fmt.Println("0. Exit")
		fmt.Println()
		fmt.Printf("Choose scenario (0-%d): ", len(list))
		line, err := in.ReadString('\n')
		choice, convErr := strconv.Atoi(strings.TrimSpace(line))
		switch {
		case convErr == nil && choice == 0:
			return
		case convErr == nil && choice >= 1 && choice <= len(list):
			fmt.Println()
			if rerr := list[choice-1].Run(); rerr != nil {
				fmt.Println("error:", rerr)
			}
			fmt.Print("\nPress Enter to return to the menu...")
			if _, perr := in.ReadString('\n'); perr != nil {
				fmt.Println()
				return
			}
			fmt.Println()
		case err != nil:
			fmt.Println()
			return
		default:
			fmt.Printf("\nInvalid choice %q, enter a number from 0 to %d.\n\n", strings.TrimSpace(line), len(list))
		}
	}
}
