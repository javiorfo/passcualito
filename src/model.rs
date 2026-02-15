use serde::{Deserialize, Serialize};

use crate::password::{PassOptions, Password};

#[derive(Debug, Serialize, Deserialize)]
pub struct Entry {
    pub name: String,
    pub password: String,
    pub info: String,
}

impl Entry {
    pub fn new(name: &str, password: &Option<String>, info: &Option<String>) -> Self {
        let password = match password {
            Some(p) => p.clone(),
            _ => Password::generate_random_password(PassOptions::Default),
        };

        Self {
            name: name.to_string(),
            password,
            info: info.as_ref().map_or(String::new(), |value| value.clone()),
        }
    }

    pub fn print(&self, is_end: bool) {
        println!("│");
        println!("├── \x1B[1mname:\x1B[0m     {}", self.name);
        println!("├── \x1B[1mpassword:\x1B[0m {}", self.password);
        if is_end {
            println!("└── \x1B[1minfo:\x1B[0m     {}", self.info);
        } else {
            println!("├── \x1B[1minfo:\x1B[0m     {}", self.info);
        }
    }

    pub fn set_password(&mut self, password: String) {
        self.password = password;
    }

    pub fn set_info(&mut self, info: String) {
        self.info = info;
    }
}
