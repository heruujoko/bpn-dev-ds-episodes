package scenarios

import (
	"fmt"
	"time"

	"rate-limit-demo/limiter"
	"rate-limit-demo/utils"
)

type KeyChoiceScenario struct{}

func (KeyChoiceScenario) Name() string { return "Choosing the key: IP vs user vs account" }
func (KeyChoiceScenario) Description() string {
	return "The key decides who shares a budget"
}

// allowByUserID shows the login bug: unauthenticated requests have no user ID,
// so the limiter is skipped entirely.
func allowByUserID(l *limiter.FixedWindow, userID string) bool {
	if userID == "" {
		return true
	}
	ok, _ := l.Allow(userID)
	return ok
}

// allowLoginAttempt applies TWO separate limits: one per target account (stops guessing
// one account from many IPs) and one per IP (stops one IP spraying many accounts).
// A single combined key "login:<account>:<ip>" would give an attacker a fresh budget
// for every new IP, so the account would be unprotected again.
func allowLoginAttempt(l *limiter.FixedWindow, account, ip string) bool {
	accountOK, _ := l.Allow("login:account:" + account)
	ipOK, _ := l.Allow("login:ip:" + ip)
	return accountOK && ipOK
}

func (KeyChoiceScenario) Run() error {
	utils.PrintHeader("Scenario 4: Choosing the Key")
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	fmt.Println("1. IP-only: 20 office employees behind one NAT IP; limit 10/min per IP")
	byIP := limiter.NewFixedWindow(10, time.Minute, clock)
	allowed, blocked := 0, 0
	for range 20 {
		if ok, _ := byIP.Allow("203.0.113.7"); ok {
			allowed++
		} else {
			blocked++
		}
	}
	utils.PrintSuccess(fmt.Sprintf("Employees allowed: %d", allowed))
	utils.PrintError(fmt.Sprintf("Employees blocked: %d legitimate users", blocked))
	utils.PrintSeparator()

	fmt.Println("2. Login by user ID: attacker is not logged in, so user ID is empty")
	byUserID := limiter.NewFixedWindow(10, time.Minute, clock)
	allowed = 0
	for range 1_000 {
		if allowByUserID(byUserID, "") {
			allowed++
		}
	}
	utils.PrintCritical(fmt.Sprintf("%d password attempts for alice@example.com allowed: no user ID means no limit.", allowed))
	utils.PrintSeparator()

	fmt.Println("3. Login: separate limits per target account AND per IP, 10/min each")
	fmt.Println("   Attacker rotates through 20 IPs (botnet), all aimed at alice@example.com")
	balanced := limiter.NewFixedWindow(10, time.Minute, clock)
	allowed, blocked = 0, 0
	for i := range 20 {
		if allowLoginAttempt(balanced, "alice@example.com", fmt.Sprintf("198.51.100.%d", i+1)) {
			allowed++
		} else {
			blocked++
		}
	}
	utils.PrintSuccess(fmt.Sprintf("Attempts allowed: %d", allowed))
	utils.PrintError(fmt.Sprintf("Attempts blocked: %d by the per-account limit (new IPs don't reset it)", blocked))
	fmt.Println()
	fmt.Println("Balanced keys:")
	fmt.Println("Endpoint                 Key                    Why")
	fmt.Println("Authenticated API        user ID / API key      fair per customer, not per network")
	fmt.Println("Login                    account, and IP        two separate limits, not one combined key")
	fmt.Println("Public signup / search   IP (generous limit)    the only identity available")
	return nil
}
