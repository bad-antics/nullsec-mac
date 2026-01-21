/*
 * NullSec macOS Privilege Escalation Scanner
 * Scans for common privesc vectors on macOS
 * Compile: clang -o nullsec-privesc privesc_scan.c -framework Security
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <dirent.h>
#include <pwd.h>
#include <grp.h>

#define RED     "\033[0;31m"
#define GREEN   "\033[0;32m"
#define YELLOW  "\033[0;33m"
#define CYAN    "\033[0;36m"
#define RESET   "\033[0m"

void banner() {
    printf(CYAN);
    printf("\n╔═══════════════════════════════════════╗\n");
    printf("║   NullSec PrivEsc Scanner - macOS     ║\n");
    printf("║   Fast C-based vulnerability scan     ║\n");
    printf("╚═══════════════════════════════════════╝\n");
    printf(RESET);
}

void check_suid_binaries() {
    printf("\n[*] Checking SUID binaries...\n");
    
    char *paths[] = {"/usr/bin", "/usr/sbin", "/usr/local/bin", "/bin", "/sbin", NULL};
    
    for (int i = 0; paths[i] != NULL; i++) {
        DIR *dir = opendir(paths[i]);
        if (!dir) continue;
        
        struct dirent *entry;
        while ((entry = readdir(dir)) != NULL) {
            char fullpath[1024];
            snprintf(fullpath, sizeof(fullpath), "%s/%s", paths[i], entry->d_name);
            
            struct stat st;
            if (stat(fullpath, &st) == 0) {
                if (st.st_mode & S_ISUID) {
                    printf(YELLOW "    [SUID] %s\n" RESET, fullpath);
                }
                if (st.st_mode & S_ISGID) {
                    printf(YELLOW "    [SGID] %s\n" RESET, fullpath);
                }
            }
        }
        closedir(dir);
    }
}

void check_writable_paths() {
    printf("\n[*] Checking writable paths in PATH...\n");
    
    char *path = getenv("PATH");
    if (!path) return;
    
    char *path_copy = strdup(path);
    char *token = strtok(path_copy, ":");
    
    while (token) {
        if (access(token, W_OK) == 0) {
            printf(RED "    [WRITABLE] %s\n" RESET, token);
        }
        token = strtok(NULL, ":");
    }
    free(path_copy);
}

void check_sudo_config() {
    printf("\n[*] Checking sudo configuration...\n");
    
    // Check if user can sudo without password
    int result = system("sudo -n true 2>/dev/null");
    if (result == 0) {
        printf(RED "    [!] NOPASSWD sudo available!\n" RESET);
    }
    
    // Check sudoers readable
    if (access("/etc/sudoers", R_OK) == 0) {
        printf(RED "    [!] /etc/sudoers is readable!\n" RESET);
    }
}

void check_launchd_persistence() {
    printf("\n[*] Checking LaunchDaemon/LaunchAgent persistence...\n");
    
    char *locations[] = {
        "/Library/LaunchDaemons",
        "/Library/LaunchAgents",
        "~/Library/LaunchAgents",
        NULL
    };
    
    for (int i = 0; locations[i] != NULL; i++) {
        char path[1024];
        if (locations[i][0] == '~') {
            char *home = getenv("HOME");
            snprintf(path, sizeof(path), "%s%s", home, locations[i] + 1);
        } else {
            strncpy(path, locations[i], sizeof(path));
        }
        
        if (access(path, W_OK) == 0) {
            printf(RED "    [WRITABLE] %s\n" RESET, path);
        } else if (access(path, R_OK) == 0) {
            printf(GREEN "    [EXISTS] %s\n" RESET, path);
        }
    }
}

void check_cron_jobs() {
    printf("\n[*] Checking cron jobs...\n");
    system("crontab -l 2>/dev/null | head -10");
    
    if (access("/etc/crontab", R_OK) == 0) {
        printf("    [*] System crontab:\n");
        system("cat /etc/crontab 2>/dev/null | grep -v '^#' | head -10");
    }
}

void check_ssh_keys() {
    printf("\n[*] Checking SSH keys...\n");
    
    char *home = getenv("HOME");
    char ssh_dir[256];
    snprintf(ssh_dir, sizeof(ssh_dir), "%s/.ssh", home);
    
    DIR *dir = opendir(ssh_dir);
    if (!dir) {
        printf("    No .ssh directory found\n");
        return;
    }
    
    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        if (entry->d_name[0] == '.') continue;
        
        char fullpath[512];
        snprintf(fullpath, sizeof(fullpath), "%s/%s", ssh_dir, entry->d_name);
        
        struct stat st;
        if (stat(fullpath, &st) == 0) {
            printf("    %s (mode: %o)\n", entry->d_name, st.st_mode & 0777);
        }
    }
    closedir(dir);
}

void check_env_vars() {
    printf("\n[*] Checking sensitive environment variables...\n");
    
    char *sensitive[] = {"AWS_", "API_KEY", "SECRET", "PASSWORD", "TOKEN", "PRIVATE", NULL};
    
    extern char **environ;
    for (char **env = environ; *env; env++) {
        for (int i = 0; sensitive[i]; i++) {
            if (strstr(*env, sensitive[i])) {
                // Mask the value
                char *eq = strchr(*env, '=');
                if (eq) {
                    int name_len = eq - *env;
                    printf(YELLOW "    %.*s=***REDACTED***\n" RESET, name_len, *env);
                }
                break;
            }
        }
    }
}

void check_sip_status() {
    printf("\n[*] Checking System Integrity Protection (SIP)...\n");
    
    int result = system("csrutil status 2>/dev/null | head -1");
    if (result != 0) {
        printf("    Unable to check SIP status\n");
    }
}

int main(int argc, char *argv[]) {
    banner();
    
    printf("\n[*] Running as: %s (UID: %d, GID: %d)\n", 
           getenv("USER"), getuid(), getgid());
    
    check_sip_status();
    check_sudo_config();
    check_suid_binaries();
    check_writable_paths();
    check_launchd_persistence();
    check_cron_jobs();
    check_ssh_keys();
    check_env_vars();
    
    printf("\n" GREEN "[✓] Privilege escalation scan complete\n" RESET);
    
    return 0;
}
