use std::collections::HashMap;

use serde::{Deserialize, Serialize};

use super::encrypt::ITEM_SEPARATOR;
use super::password::{PassOptions, generate_random_password};

#[derive(Serialize, Deserialize, Debug, Default)]
pub struct Input {
    pub name: String,
    pub password: String,
    pub info: Option<String>,
}

impl Input {
    pub fn new(name: String, password: Option<String>, info: Option<String>) -> Self {
        let password = match password {
            Some(p) => p,
            _ => generate_random_password(PassOptions::Default),
        };

        Self {
            name,
            password,
            info,
        }
    }

    pub fn print(self, is_end: bool) {
        println!("│");
        println!("├── \033[1mname:\033[0m    {}", self.name);
        println!("├── \033[1mpassword:\033[0m{}", self.password);
        if is_end {
            println!(
                "└── \033[1minfo:\033[0m    {}",
                self.info.unwrap_or(String::from(""))
            );
        } else {
            println!(
                "├── \033[1minfo:\033[0m    {}",
                self.info.unwrap_or(String::from(""))
            );
        }
    }

    pub fn is_name_taken(content: &str, name: &str) -> bool {
        let mut items = content.split(ITEM_SEPARATOR);
        items.any(|json_data| {
            let data: Input = serde_json::from_str(json_data).unwrap_or_default();
            data.name == name
        })
    }

    pub fn is_name_matched(&self, input_name: &str) -> bool {
        let name_lower = self.name.to_lowercase();
        let input_name_lower = input_name.to_lowercase();

        if input_name_lower.len() == 1 {
            name_lower.contains(&input_name_lower)
        } else {
            name_lower.starts_with(&input_name_lower)
        }
    }

    pub fn string_to_input_list(content: &str) -> Vec<Input> {
        content
            .split(ITEM_SEPARATOR)
            .map(|item| serde_json::from_str(item).unwrap_or_default())
            .collect()
    }

    fn get_input_list_from_json_file(
        file_path: &str,
    ) -> Result<Vec<Input>, Box<dyn std::error::Error>> {
        let json_data = std::fs::read_to_string(file_path)?;

        let data_vec: Vec<Input> = serde_json::from_str(&json_data)?;

        // TODO get_repeated_names

        Ok(data_vec)
    }

    fn get_repeated_names(data_slice: &[Input]) -> Option<String> {
        let mut counts: HashMap<&str, usize> = HashMap::new();

        for data in data_slice {
            *counts.entry(&data.name).or_insert(0) += 1;
        }

        let repeated: Vec<String> = counts
            .iter()
            .filter_map(|(&name, &count)| {
                if count > 1 {
                    Some(name.to_string())
                } else {
                    None
                }
            })
            .collect();

        if !repeated.is_empty() {
            Some(repeated.join(", "))
        } else {
            None
        }
    }
}
