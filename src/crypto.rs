use aes_gcm::{
    Aes256Gcm, Nonce,
    aead::{Aead, KeyInit, OsRng},
};
use argon2::{
    Argon2,
    password_hash::{PasswordHasher, SaltString},
};
use rand_core::RngCore;

pub fn derive_key(master_password: &str, salt: &[u8]) -> [u8; 32] {
    let argon2 = Argon2::default();
    let salt = SaltString::encode_b64(salt).unwrap();
    let password_hash = argon2
        .hash_password(master_password.as_bytes(), &salt)
        .unwrap();

    let mut key = [0u8; 32];
    key.copy_from_slice(password_hash.hash.unwrap().as_bytes());
    key
}

pub fn generate_salt() -> [u8; 16] {
    let mut salt = [1u8; 16];
    OsRng.fill_bytes(&mut salt);
    salt
}

pub fn encrypt(key: &[u8; 32], data: &[u8]) -> anyhow::Result<(Vec<u8>, Vec<u8>)> {
    let cipher = Aes256Gcm::new_from_slice(key).map_err(|e| anyhow::anyhow!(e))?;
    let mut nonce_bytes = [0u8; 12];
    OsRng.fill_bytes(&mut nonce_bytes);
    let nonce = Nonce::from_slice(&nonce_bytes);

    let ciphertext = cipher
        .encrypt(nonce, data)
        .map_err(|e| anyhow::anyhow!(e))?;
    Ok((ciphertext, nonce_bytes.to_vec()))
}

pub fn decrypt(key: &[u8; 32], ciphertext: &[u8], nonce_bytes: &[u8]) -> anyhow::Result<Vec<u8>> {
    let cipher = Aes256Gcm::new_from_slice(key).map_err(|e| anyhow::anyhow!(e))?;
    let nonce = Nonce::from_slice(nonce_bytes);

    let plaintext = cipher
        .decrypt(nonce, ciphertext.as_ref())
        .map_err(|e| anyhow::anyhow!(e))?;
    Ok(plaintext)
}
