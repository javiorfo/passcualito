use std::{
    io::Write,
    process::{Command, Stdio},
};

use crate::{model::Entry, password::Charset};
use crate::{password::Password, storage::Storage};
use anyhow::bail;
use clap::{Parser, Subcommand};
use rpassword::prompt_password_stdout;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    Add {
        name: String,

        #[arg(short, long)]
        #[arg(help = "Leave empty to generate a strong password")]
        password: Option<String>,

        #[arg(short, long)]
        #[arg(help = "Optional info (user, url, etc).")]
        info: Option<String>,
    },
    Edit {
        name: String,

        #[arg(short, long)]
        #[arg(help = "Edit password")]
        password: Option<String>,

        #[arg(short, long)]
        #[arg(help = "Edit info")]
        info: Option<String>,
    },
    List {
        name: Option<String>,
    },
    Copy {
        entry: String,
    },
    Password {
        digits: usize,

        #[arg(short, long)]
        #[arg(
            help = "Charset could be a, c, n, an or anc. Alphabetic, numeric, alphanumeric, alphanumeric and capital letters, respectively. If left empty the charset also include special characters."
        )]
        charset: Option<String>,
    },
    Export,
    Import {
        path: String,
    },
    Remove {
        name: String,
    },
}

pub fn run_cli() -> anyhow::Result<()> {
    let args = Cli::parse();

    println!("\x1B[1m  Passcualito\x1B[0m");
    let master_password =
        Password::validate_master_password(prompt_password_stdout("  Master Password: ")?)?;

    let mut storage = match Storage::load_store(&master_password) {
        Ok(s) => s,
        Err(e) => {
            if e.downcast_ref::<aes_gcm::Error>().is_some() {
                bail!("Incorrect Password!")
            }
            eprintln!("󰸞  Master password and storage have been created");
            Storage::default()
        }
    };

    match &args.command {
        Commands::Add {
            name,
            password,
            info,
        } => {
            if storage.get_entry(name).is_some() {
                bail!("Entry '{name}' already exists")
            }

            storage.entries.push(Entry::new(name, password, info));
            println!("󰸞  Entry \x1B[1m{name}\x1B[0m created");
        }
        Commands::Edit {
            name,
            password,
            info,
        } => match storage.get_entry(name) {
            Some(entry) => {
                if let Some(pass) = password {
                    entry.set_password(pass.to_string());
                }
                if let Some(info) = info {
                    entry.set_info(info.to_string());
                }

                println!("󰸞  Entry \x1B[1m{name}\x1B[0m has been edited")
            }
            _ => bail!("Entry '{name}' does not exist!"),
        },
        Commands::List { name } => match name {
            Some(name) => {
                println!("\x1B[1m󰪶  Matches\x1B[0m");
                let filtered_entries = storage
                    .entries
                    .iter()
                    .filter(|&e| e.name.contains(name))
                    .collect::<Vec<_>>();

                let size = filtered_entries.len();
                for (i, entry) in filtered_entries.iter().enumerate() {
                    if i == size - 1 {
                        entry.print(true);
                    } else {
                        entry.print(false);
                    }
                }

                if size == 0 {
                    println!("󰮗  Entry name \x1B[1m{name}\x1B[0m not found.");
                }
            }
            None => {
                if storage.entries.is_empty() {
                    println!("󱀰  No data stored yet.");
                } else {
                    println!("\x1B[1m󰪶  Storage\x1B[0m");
                    let size = storage.entries.len();
                    for (i, entry) in storage.entries.iter().enumerate() {
                        if i == size - 1 {
                            entry.print(true);
                        } else {
                            entry.print(false);
                        }
                    }
                }
            }
        },
        Commands::Copy { entry } => match storage.get_entry(entry) {
            Some(entry) => {
                let mut child = if std::env::var("WAYLAND_DISPLAY").is_ok() {
                    Command::new("wl-copy").stdin(Stdio::piped()).spawn()?
                } else {
                    Command::new("xclip")
                        .args(["-selection", "clipboard"])
                        .stdin(Stdio::piped())
                        .spawn()?
                };

                {
                    let stdin = child.stdin.as_mut().expect("Failed to open stdin");
                    stdin.write_all(entry.password.as_bytes())?;
                }

                let _ = child.wait()?;

                println!(
                    "󰢨  Password of entry \x1B[1m{}\x1B[0m copied to clipboard",
                    entry.name
                );
            }
            None => bail!("Entry '{entry}' not found!"),
        },
        Commands::Password { digits, charset } => {
            let password = Charset::password(*digits, charset);
            println!("  Password generated: {password}");
        }
        Commands::Export => {
            storage.export()?;
            println!("󰸞  Data exported to \x1B[1m pascualito.json\x1B[0m");
        }
        Commands::Import { path } => {
            storage.import_from_file(path)?;
            println!("󰸞  \x1B[1m{path}\x1B[0m has been imported.")
        }
        Commands::Remove { name } => match storage.get_entry(name) {
            Some(_) => {
                let position = storage
                    .entries
                    .iter()
                    .position(|e| e.name == *name)
                    .unwrap();

                storage.entries.remove(position);
                println!("󰸞  Entry \x1B[1m{name}\x1B[0m has been removed.");
            }
            None => bail!("Entry '{name}' does not exist!"),
        },
    }

    storage.save_store(&master_password)
}
