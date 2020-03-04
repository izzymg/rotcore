use ring::hmac;

pub type Key = hmac::Key;

pub fn make_key(secret: &[u8]) -> Key {
    hmac::Key::new(hmac::HMAC_SHA256, secret)
}

/// Verify's that `mac` matches the hash of `body` + `key`.
pub fn verify_hash(body: &[u8], mac: &[u8], key: &Key) -> Result<bool, &'static str> {
    if body.len() < 1 {
        return Err("Body is required");
    }
    if mac.len() < 1 {
        return Err("MAC is required");
    }

    match hmac::verify(&key, body, mac) {
        Ok(_) => Ok(true),
        Err(_) => Ok(false),
    }
}

#[cfg(test)]
mod test {
    use ring::hmac;
    #[test]
    fn test_verify_hash() {
        let key = hmac::Key::new(hmac::HMAC_SHA256, "AAAAAAAAAIAMSCREAMING".as_bytes());
        let bad_key = hmac::Key::new(hmac::HMAC_SHA256, "?hhhh?".as_bytes());

        // Try some garbage inputs
        let v = super::verify_hash(b"", b"something", &key);
        assert!(v.is_err());
        let v = super::verify_hash(b"something", b"", &key);
        assert!(v.is_err());
        let v = super::verify_hash(b"something", b"", &key);
        assert!(v.is_err());

        // Test some garbage hashes
        let v = super::verify_hash(b"something", b"something", &key);
        assert!(!v.is_err());

        let v = super::verify_hash(b"AAA", b"AAA", &key);
        assert!(!v.is_err());
        assert!(v.unwrap() == false);

        // Test a good hash
        let data = b"some data related to something";
        let mac = hmac::sign(&key, data);
        let v = super::verify_hash(data, mac.as_ref(), &key);
        assert!(!v.is_err());
        assert!(v.unwrap() == true);

        // Test bad key
        let mac = hmac::sign(&bad_key, data);
        let v = super::verify_hash(data, mac.as_ref(), &key);
        assert!(!v.is_err());
        assert!(v.unwrap() == false);
    }
}
