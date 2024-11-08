use aes::Aes256;
use block_modes::{BlockMode, Cbc};
use rand::Rng;
use ring::pbkdf2;
use rpassword::read_password;
use std::num::NonZeroU32;
use std::fs::File;
use std::io::{Read, Write};
use std::path::Path;

// type Aes256Cbc = Cbc<Aes256, rand::rngs::OsRng>;
type Aes256Cbc = Cbc<Aes256, block_modes::block_padding::Pkcs7>;

const SALT: &[u8] = b"unique_salt"; // Use a unique salt for your application
const ITERATIONS: u32 = 100_000;

fn derive_key(password: &str) -> Vec<u8> {
    let mut key = [0u8; 32]; // AES-256 key size
    pbkdf2::derive(
        pbkdf2::PBKDF2_HMAC_SHA256,
        NonZeroU32::new(ITERATIONS).unwrap(),
        SALT,
        password.as_bytes(),
        &mut key,
    );
    key.to_vec()
}

fn encrypt_file<P: AsRef<Path>>(path: P, password: &str) -> std::io::Result<()> {
    let key = derive_key(password);
    let iv: [u8; 16] = rand::thread_rng().gen(); // Generate a random IV

    let mut file = File::open(&path)?;
    let mut data = Vec::new();
    file.read_to_end(&mut data)?;

    let cipher = Aes256Cbc::new_from_slices(&key, &iv).unwrap();
    let encrypted_data = cipher.encrypt_vec(&data);

    let mut output = File::create(path)?;
    output.write_all(&iv)?;
    output.write_all(&encrypted_data)?;

    Ok(())
}

fn decrypt_file<P: AsRef<Path>>(path: P, password: &str) -> std::io::Result<Vec<u8>> {
    let key = derive_key(password);
    let mut file = File::open(path)?;
    
    let mut iv = [0u8; 16];
    file.read_exact(&mut iv)?;

    let mut encrypted_data = Vec::new();
    file.read_to_end(&mut encrypted_data)?;

    let cipher = Aes256Cbc::new_from_slices(&key, &iv).unwrap();
    let decrypted_data = cipher.decrypt_vec(&encrypted_data).unwrap();

    Ok(decrypted_data)
}

// funcion crear store file ".config/passcualito/passwords
// fn main() {
//     let password = "123"; // Use a secure password
//     let file_path = "/home/javier/dev/rust/passman/file.txt";
// 
//     // Encrypt the file
//     if let Err(e) = encrypt_file(file_path, password) {
//         eprintln!("Error encrypting file: {}", e);
//     }
// 
//     // Decrypt the file
//     match decrypt_file(file_path, password) {
//         Ok(data) => {
//             println!("Decrypted data: {:?}", String::from_utf8(data).unwrap());
//         }
//         Err(e) => {
//             eprintln!("Error decrypting file: {}", e);
//         }
//     }
// }

// funcion login leer
fn main() {
    let file_path = "/home/javier/dev/rust/passman/file.txt";

    println!("Enter password for encryption/decryption:");
    let password = read_password().expect("Failed to read password");
//     if let Err(e) = encrypt_file(file_path, &password) {
//         eprintln!("Error encrypting file: {}", e);
//     }

    match decrypt_file(file_path, &password) {
        Ok(data) => {
            println!("Decrypted data: {:?}", String::from_utf8(data));
        }
        Err(e) => {
            eprintln!("Error decrypting file: {}", e);
        }
    }
}
