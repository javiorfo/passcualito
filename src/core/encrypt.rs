use aes_gcm::{
    aead::{Aead, AeadCore, KeyInit, OsRng},
    Aes256Gcm, Nonce, Key
};
use serde_json::Value;
use std::{
    fs::{self, File, OpenOptions},
    io::{self, Read, Write},
    path::{Path, PathBuf},
};

pub const ITEM_SEPARATOR: &str = "|";
const BACKUP_FILENAME: &str = "passc_backup.json";
const EXPORT_FILENAME: &str = "passc_export.json";

pub struct Encryptor {
    master_password: String,
    file_path: PathBuf,
}

impl Encryptor {
    fn new(master_password: String, file_path: impl AsRef<Path>) -> Self {
        Self {
            master_password,
            file_path: file_path.as_ref().to_path_buf(),
        }
    }

    fn delete_content(&self) -> io::Result<()> {
        File::create(&self.file_path)?;
        Ok(())
    }

    fn encrypt_text(&self, text: &str, is_append: bool) -> io::Result<()> {
        let mut final_text = text.to_string();
        
        // Handle existing content if appending
        if is_append && self.file_path.exists() && fs::metadata(&self.file_path)?.len() > 0 {
            if let Ok(decrypted) = self.read_encrypted_text() {
                final_text.push_str(ITEM_SEPARATOR);
                final_text.push_str(&decrypted);
            }
        }

        // Create cipher (256-bit key derived from password)
        let key = Key::<Aes256Gcm>::from_slice(self.master_password.as_bytes()); // Note: In real usage, use proper key derivation!
        let cipher = Aes256Gcm::new(key);

        // Generate nonce and encrypt
        let nonce = Aes256Gcm::generate_nonce(&mut OsRng); // 96-bit nonce
        let ciphertext = cipher
            .encrypt(&nonce, final_text.as_bytes()).unwrap();
//             .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

        // Write nonce + ciphertext
        let mut file = OpenOptions::new()
            .write(true)
            .create(true)
            .truncate(!is_append)
            .open(&self.file_path)?;
        
        file.write_all(&nonce)?;
        file.write_all(&ciphertext)?;

        Ok(())
    }

    fn read_encrypted_text(&self) -> io::Result<String> {
        let mut file = File::open(&self.file_path)?;
        let mut data = Vec::new();
        file.read_to_end(&mut data)?;

        if data.is_empty() {
            return Err(io::Error::new(io::ErrorKind::Other, "empty file"));
        }

        // Split nonce and ciphertext
        let cipher = Aes256Gcm::new(Key::<Aes256Gcm>::from_slice(self.master_password.as_bytes()));
        let nonce_size = 12; // AES-GCM standard nonce size
        
        if data.len() < nonce_size {
            return Err(io::Error::new(io::ErrorKind::InvalidData, "file too short"));
        }

        let (nonce, ciphertext) = data.split_at(nonce_size);
        let nonce = Nonce::from_slice(nonce);

        let plaintext = cipher
            .decrypt(nonce, ciphertext)
            .unwrap();
//             .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

        String::from_utf8(plaintext).map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))
    }
}

// Export functionality
fn export_to_file(content: &str) -> io::Result<()> {
    let items: Vec<&str> = content.split(ITEM_SEPARATOR).collect();
    let mut file = File::create(EXPORT_FILENAME)?;

    file.write_all(b"[")?;
    
    for (i, item) in items.iter().enumerate() {
        // Parse and pretty-print JSON
        let value: Value = serde_json::from_str(item).map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        let pretty = serde_json::to_string_pretty(&value)?;
        
        file.write_all(pretty.as_bytes())?;
        
        if i != items.len() - 1 {
            file.write_all(b",")?;
        }
    }
    
    file.write_all(b"]")?;
    Ok(())
}

// Backup functionality
fn make_backup() -> io::Result<()> {
    let home_dir = dirs::home_dir().ok_or_else(|| io::Error::new(io::ErrorKind::NotFound, "home directory not found"))?;
    
    let src = home_dir.join("passc_store.enc"); // Adjust filename as needed
    let backup_dir = home_dir.join(".passc_backups");
    
    fs::create_dir_all(&backup_dir)?;
    
    let dst = backup_dir.join(BACKUP_FILENAME);
    fs::copy(src, dst)?;
    
    Ok(())
}
