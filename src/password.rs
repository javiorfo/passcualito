use anyhow::{Ok, bail};
use rand::{Rng, rng};

const ALPHABETIC: &str = "abcdefghijklmnopqrstuvwxyz";
const CAPITAL: &str = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
const NUMERIC: &str = "0123456789";
const ALPHA_NUMERIC: &str = "abcdefghijklmnopqrstuvwxyz0123456789";
const ALPHA_NUMERIC_CAPITAL: &str =
    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
const DEFAULT_CHARSET: &str =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*?";

pub enum PassOptions {
    LengthCharset(usize, Charset),
    Length(usize),
    Default,
}

impl PassOptions {
    pub const DEFAULT_PASSWORD_LENGTH: usize = 20;

    fn length_and_charset(&self) -> (usize, &str) {
        match self {
            PassOptions::LengthCharset(s, c) => (*s, c.into()),
            PassOptions::Length(s) => (*s, DEFAULT_CHARSET),
            PassOptions::Default => (PassOptions::DEFAULT_PASSWORD_LENGTH, DEFAULT_CHARSET),
        }
    }
}

pub struct Password;

impl Password {
    pub fn validate_master_password(password: String) -> anyhow::Result<String> {
        if password.len() > 5 {
            Ok(password)
        } else {
            bail!("Master Password must have at least 6 characters")
        }
    }

    pub fn generate_random_password(pass_options: PassOptions) -> String {
        let (length, charset) = pass_options.length_and_charset();
        let charset_bytes = charset.as_bytes();
        let mut rng = rng();

        let password: String = (0..length)
            .map(|_| {
                let idx = rng.random_range(0..charset_bytes.len());
                charset_bytes[idx] as char
            })
            .collect();

        password
    }
}

pub enum Charset {
    Numeric,
    Alphabetic,
    Capital,
    AlphaNumeric,
    AlphaNumericCapital,
    Default,
}

impl Charset {
    pub fn password(length: usize, charset: &Option<String>) -> String {
        match charset {
            Some(charset) => {
                let charset: Charset = charset.into();
                Password::generate_random_password(PassOptions::LengthCharset(length, charset))
            }
            None => Password::generate_random_password(PassOptions::Length(length)),
        }
    }
}

impl From<&String> for Charset {
    fn from(value: &String) -> Self {
        match value.as_str() {
            "a" => Self::Alphabetic,
            "c" => Self::Capital,
            "n" => Self::Numeric,
            "an" => Self::AlphaNumeric,
            "anc" => Self::AlphaNumericCapital,
            _ => Self::Default,
        }
    }
}

impl From<&Charset> for &str {
    fn from(value: &Charset) -> Self {
        match value {
            Charset::Alphabetic => ALPHABETIC,
            Charset::AlphaNumeric => ALPHA_NUMERIC,
            Charset::AlphaNumericCapital => ALPHA_NUMERIC_CAPITAL,
            Charset::Capital => CAPITAL,
            Charset::Numeric => NUMERIC,
            Charset::Default => DEFAULT_CHARSET,
        }
    }
}
