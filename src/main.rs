use clap::Parser;
use clap::Subcommand;
use commands::add::Add;

mod commands;
mod core;

#[derive(Parser, Debug)]
#[command(
    name = "passc",
    version = "1.0",
    about = "A CLI tool for managing entries in a store"
)]
struct Passcualito {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    Add(Add),
}

fn main() {
    let passcualito = Passcualito::parse();

    match passcualito.command {
        Commands::Add(args) => args.handle(),
    }
}
