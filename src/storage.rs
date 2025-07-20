use anyhow::bail;
use serde::{Deserialize, Serialize};

use crate::crypto::{self, derive_key};
use crate::model::Entry;
use std::path::{Path, PathBuf};
use std::{env, fs};

#[derive(Debug, Serialize, Deserialize, Default)]
pub struct Storage {
    pub entries: Vec<Entry>,
}

impl Storage {
    const PASSWORD_FILE: &str = "passwords.dat";
    const EXPORT_FILE: &str = "passcualito.json";

    pub fn get_entry(&mut self, name: &str) -> Option<&mut Entry> {
        self.entries.iter_mut().find(|data| data.name == name)
    }

    pub fn load_store(master_password: &str) -> anyhow::Result<Storage> {
        let path_buf = Self::create_password_dir()?.join(Self::PASSWORD_FILE);
        let encrypted_data = fs::read(path_buf)?;

        let salt = &encrypted_data[0..16];
        let encrypted_store_data = &encrypted_data[16..];

        let key = derive_key(master_password, salt);

        if encrypted_store_data.len() < 12 {
            bail!("Invalid file format: Not enough data for nonce")
        }
        let nonce = &encrypted_store_data[0..12];
        let ciphertext = &encrypted_store_data[12..];

        let decrypted_bytes = crypto::decrypt(&key, ciphertext, nonce)?;
        let mut storage: Storage = serde_json::from_slice(&decrypted_bytes)?;
        storage.entries.sort_by_key(|e| e.name.clone());
        Ok(storage)
    }

    pub fn save_store(&self, master_password: &str) -> anyhow::Result<()> {
        let path_buf = Self::create_password_dir()?.join(Self::PASSWORD_FILE);

        let serialized_store = serde_json::to_vec(self)?;

        let salt = crypto::generate_salt();
        let key = derive_key(master_password, &salt);

        let (encrypted_bytes, nonce_bytes) = crypto::encrypt(&key, &serialized_store).unwrap();

        let mut data = Vec::new();
        data.extend_from_slice(&salt);
        data.extend_from_slice(&nonce_bytes);
        data.extend_from_slice(&encrypted_bytes);

        fs::write(path_buf, &data)?;
        Ok(())
    }

    pub fn export(&self) -> anyhow::Result<()> {
        let data = serde_json::to_string_pretty(&self.entries)?;
        fs::write(Self::EXPORT_FILE, &data)?;
        Ok(())
    }

    pub fn import_from_file(&mut self, path: &str) -> anyhow::Result<()> {
        let data = fs::read(path)?;
        let mut entries: Vec<Entry> = serde_json::from_slice(&data)?;

        entries.sort_by_key(|e| e.name.clone());
        Self::validated_repeated(&entries)?;

        if let Some(repeated) = self
            .entries
            .iter()
            .find(|&e| entries.iter().any(|e2| e2.name == e.name))
        {
            bail!(
                "Entry with name '{}' already exists in the current storage!",
                repeated.name
            )
        }

        self.entries.extend(entries);

        Ok(())
    }

    fn validated_repeated(entries: &[Entry]) -> anyhow::Result<()> {
        for i in 1..entries.len() {
            if entries[i - 1].name == entries[i].name {
                bail!(
                    "Entry with name '{}' is repeated in the file",
                    entries[i].name
                )
            }
        }
        Ok(())
    }

    fn create_password_dir() -> anyhow::Result<PathBuf> {
        let folder_path = Path::new(&env::var("HOME")?).join(".passcualito");
        fs::create_dir_all(&folder_path)?;
        Ok(folder_path)
    }
}
