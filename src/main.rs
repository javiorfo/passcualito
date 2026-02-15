mod cli;
mod crypto;
mod model;
mod password;
mod storage;

fn main() {
    if let Err(e) = cli::run_cli() {
        eprintln!("\x1b[31m  {e} \x1b[0m");
        std::process::exit(1);
    }
}
