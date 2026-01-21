// NullSec macOS Network Scanner
// Fast ARP/port scanning for local networks
// Build: go build -o nullsec-netscan netscan.go
package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Host struct {
	IP       string
	MAC      string
	Hostname string
	Ports    []int
}

func banner() {
	fmt.Println(`
╔═══════════════════════════════════════╗
║   NullSec Network Scanner - macOS     ║
║   Fast concurrent scanning in Go      ║
╚═══════════════════════════════════════╝`)
}

func getLocalInterfaces() {
	ifaces, _ := net.Interfaces()
	fmt.Println("\n[*] Network Interfaces:")
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					fmt.Printf("    %s: %s\n", iface.Name, ipnet.IP.String())
				}
			}
		}
	}
}

func arpScan() []Host {
	var hosts []Host
	cmd := exec.Command("arp", "-a")
	out, _ := cmd.Output()

	re := regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\) at ([0-9a-f:]+)`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	for _, match := range matches {
		if len(match) >= 3 {
			hosts = append(hosts, Host{
				IP:  match[1],
				MAC: match[2],
			})
		}
	}
	return hosts
}

func scanPort(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func scanPorts(ip string, ports []int, timeout time.Duration) []int {
	var open []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 100) // Limit concurrent connections

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if scanPort(ip, p, timeout) {
				mu.Lock()
				open = append(open, p)
				mu.Unlock()
			}
		}(port)
	}
	wg.Wait()
	sort.Ints(open)
	return open
}

func resolveHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func main() {
	banner()

	target := flag.String("target", "", "Target IP or range (e.g., 192.168.1.1 or 192.168.1.0/24)")
	ports := flag.String("ports", "22,80,443,8080", "Ports to scan (comma-separated)")
	timeout := flag.Int("timeout", 500, "Timeout in milliseconds")
	arp := flag.Bool("arp", false, "Perform ARP scan")
	flag.Parse()

	getLocalInterfaces()

	if *arp {
		fmt.Println("\n[*] ARP Scan Results:")
		hosts := arpScan()
		for _, h := range hosts {
			hostname := resolveHostname(h.IP)
			fmt.Printf("    %s (%s) %s\n", h.IP, h.MAC, hostname)
		}
		fmt.Printf("\n[✓] Found %d hosts\n", len(hosts))
	}

	if *target != "" {
		fmt.Printf("\n[*] Scanning %s...\n", *target)

		var portList []int
		for _, p := range strings.Split(*ports, ",") {
			var port int
			fmt.Sscanf(strings.TrimSpace(p), "%d", &port)
			if port > 0 {
				portList = append(portList, port)
			}
		}

		open := scanPorts(*target, portList, time.Duration(*timeout)*time.Millisecond)
		if len(open) > 0 {
			fmt.Printf("    Open ports: %v\n", open)
		} else {
			fmt.Println("    No open ports found")
		}
	}

	if *target == "" && !*arp {
		fmt.Println("\n[!] Usage: nullsec-netscan [-arp] [-target IP] [-ports 22,80,443]")
		flag.PrintDefaults()
	}
}
