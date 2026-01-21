// NullSec macOS Process Injector
// Inject code into running processes
// Build: go build -o nullsec-inject inject.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func banner() {
	fmt.Println(`
╔═══════════════════════════════════════╗
║   NullSec Process Injector - macOS    ║
║   Fast memory injection in Go         ║
╚═══════════════════════════════════════╝`)
}

func listProcesses() {
	cmd := exec.Command("ps", "aux")
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")

	fmt.Println("\n[*] Running processes:")
	fmt.Println(lines[0]) // Header
	for i, line := range lines[1:] {
		if i > 20 {
			fmt.Printf("    ... and %d more\n", len(lines)-22)
			break
		}
		if line != "" {
			fmt.Println(line)
		}
	}
}

func getProcessInfo(pid int) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,ppid,user,%cpu,%mem,command")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("[!] Process %d not found\n", pid)
		return
	}
	fmt.Printf("\n[*] Process %d info:\n%s", pid, string(out))
}

func injectDylib(pid int, dylib string) error {
	fmt.Printf("[*] Attempting to inject %s into PID %d\n", dylib, pid)

	// Check if dylib exists
	if _, err := os.Stat(dylib); os.IsNotExist(err) {
		return fmt.Errorf("dylib not found: %s", dylib)
	}

	// Use DYLD_INSERT_LIBRARIES technique
	env := fmt.Sprintf("DYLD_INSERT_LIBRARIES=%s", dylib)
	fmt.Printf("[*] Setting environment: %s\n", env)

	// Attach to process using ptrace equivalent
	fmt.Printf("[!] Direct injection requires SIP disabled\n")
	fmt.Printf("[*] Alternative: Use frida-inject or insert_dylib\n")

	return nil
}

func suspendProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGSTOP)
}

func resumeProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGCONT)
}

func main() {
	banner()

	pid := flag.Int("pid", 0, "Target process ID")
	dylib := flag.String("dylib", "", "Dylib to inject")
	list := flag.Bool("list", false, "List running processes")
	info := flag.Bool("info", false, "Get process info")
	suspend := flag.Bool("suspend", false, "Suspend process")
	resume := flag.Bool("resume", false, "Resume process")
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Println("[!] Warning: Root required for injection")
	}

	if *list {
		listProcesses()
		return
	}

	if *pid == 0 {
		fmt.Println("[!] Usage: nullsec-inject -pid <PID> [-dylib <path>] [-info] [-suspend] [-resume]")
		flag.PrintDefaults()
		return
	}

	if *info {
		getProcessInfo(*pid)
	}

	if *suspend {
		if err := suspendProcess(*pid); err != nil {
			fmt.Printf("[!] Failed to suspend: %v\n", err)
		} else {
			fmt.Printf("[✓] Process %d suspended\n", *pid)
		}
	}

	if *resume {
		if err := resumeProcess(*pid); err != nil {
			fmt.Printf("[!] Failed to resume: %v\n", err)
		} else {
			fmt.Printf("[✓] Process %d resumed\n", *pid)
		}
	}

	if *dylib != "" {
		if err := injectDylib(*pid, *dylib); err != nil {
			fmt.Printf("[!] Injection failed: %v\n", err)
		}
	}
}
