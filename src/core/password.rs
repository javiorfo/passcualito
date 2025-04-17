use rand::{Rng, rng};

pub enum PassOptions<'a> {
    LengthCharset(usize, &'a str),
    Charset(&'a str),
    Default,
}

impl<'a> PassOptions<'a> {
    pub const DEFAULT_CHARSET: &'a str =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";

    pub const DEFAULT_PASSWORD_LENGTH: usize = 20;

    fn length_and_charset(self) -> (usize, &'a str) {
        match self {
            PassOptions::LengthCharset(s, c) => (s, c),
            PassOptions::Charset(c) => (PassOptions::DEFAULT_PASSWORD_LENGTH, c),
            PassOptions::Default => (
                PassOptions::DEFAULT_PASSWORD_LENGTH,
                PassOptions::DEFAULT_CHARSET,
            ),
        }
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

pub fn align_password(password: &str) -> String {
    match password.len() {
        len if len < 16 => format!("{}{}", password, "*".repeat(16 - len)),
        len if len > 16 => password[..16].to_string(),
        _ => password.to_string(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_random_password_default() {
        let password = generate_random_password(PassOptions::Default);
        assert_eq!(password.len(), 20);

        for c in password.chars() {
            assert!(PassOptions::DEFAULT_CHARSET.contains(c));
        }
    }

    #[test]
    fn test_generate_random_password_custom_charset() {
        let charset = "abc123";
        let password = generate_random_password(PassOptions::LengthCharset(50, charset));
        assert_eq!(password.len(), 50);
        for c in password.chars() {
            assert!(charset.contains(c));
        }
    }

    #[test]
    fn test_align_password_short() {
        let input = "short";
        let aligned = align_password(input);
        assert_eq!(aligned.len(), 16);
        assert!(aligned.starts_with("short"));
        assert!(aligned.ends_with(&"*".repeat(16 - input.len())));
    }

    #[test]
    fn test_align_password_exact() {
        let input = "sixteen_charstr";
        assert_eq!(input.len(), 15);
        let input = "sixteen_char_str";
        assert_eq!(input.len(), 16);
        let aligned = align_password(input);
        assert_eq!(aligned, input);
    }

    #[test]
    fn test_align_password_long() {
        let input = "thispasswordiswaytoolong";
        let aligned = align_password(input);
        assert_eq!(aligned.len(), 16);
        assert_eq!(aligned, &input[..16]);
    }
}
