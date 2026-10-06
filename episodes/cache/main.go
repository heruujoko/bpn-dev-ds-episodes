// Cache demo: an interactive menu that runs five cache scenarios.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cache-demo/scenarios"
)

// Scenario is anything the menu can run.
type Scenario interface {
	Name() string
	Description() string
	Run() error
}

func main() {
	list := []Scenario{
		scenarios.HappyPathScenario{},
		scenarios.PermissionsScenario{},
		scenarios.InventoryScenario{},
		scenarios.KeyGranularityScenario{},
		scenarios.MultiServerScenario{},
	}
	in := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("=== Cache Design System Episodes ===")
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
				return // stdin closed
			}
			fmt.Println()
		case err != nil:
			fmt.Println() // stdin closed: exit instead of looping forever
			return
		default:
			fmt.Printf("\nInvalid choice %q, enter a number from 0 to %d.\n\n", strings.TrimSpace(line), len(list))
		}
	}
}
