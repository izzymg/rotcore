// Maps allowed US characters to their respective X11 keycodes.

/// Codes for possible mouse buttons.
#[derive(Copy, Debug, Clone)]
pub enum MouseCode {
    Left = 1,
    Right = 3,
    ScrollUp = 4,
    ScrollDown = 5,
}

/// Grab a mouse code from a u8.
/// TODO: make this less ugly (remove replication of enums?)
pub fn mousecode_from_u8(i: u8) -> Result<MouseCode, &'static str> {
    match i {
        x if x == MouseCode::Left as u8 => Ok(MouseCode::Left),
        x if x == MouseCode::Right as u8  => Ok(MouseCode::Right),
        x if x == MouseCode::ScrollUp as u8 => Ok(MouseCode::ScrollUp),
        x if x == MouseCode::ScrollDown as u8 => Ok(MouseCode::ScrollDown),
        _ => Err("Invalid mouse code"),
    }
}

/// Codes for possible input event types.
pub enum InputEvent {
    KeyPress = 2,
    KeyRelease = 3,
    MousePress = 4,
    MouseRelease = 5,
}

/// Keycode for the left shift key.
pub const SHIFT_KEYCODE: u8 = 50;

/// Maps special keys to their keycodes
#[derive(Copy, Clone, Debug)]
pub enum SpecialCode {
    Backspace = 22,
    Tab = 23,
    Return = 36,
    Space = 65,
    Up = 111,
    Down = 116,
    Left = 113,
    Right = 114,
}

/// Maps strings to special codes
pub fn special_from_str(s: &str) -> Result<SpecialCode, &'static str> {
    match s {
        "backspace" => Ok(SpecialCode::Backspace),
        "space" => Ok(SpecialCode::Space),
        "tab" => Ok(SpecialCode::Tab),
        "return" => Ok(SpecialCode::Return),
        "up" => Ok(SpecialCode::Up),
        "down" => Ok(SpecialCode::Down),
        "left" => Ok(SpecialCode::Left),
        "right" => Ok(SpecialCode::Right),
        _ => Err("Invalid special code"),
    }
}

/// Returns zero (0) if not matched.
pub fn get_upper_key(key: char) -> u8 {
    match key {
        // Row 1
        '!' => 10,
        '@' => 11,
        '#' => 12,
        '$' => 13,
        '%' => 14,
        '^' => 15,
        '&' => 16,
        '*' => 17,
        '(' => 18,
        ')' => 19,
        '_' => 20,
        '+' => 21,

        // Row 2
        'Q' => 24,
        'W' => 25,
        'E' => 26,
        'R' => 27,
        'T' => 28,
        'Y' => 29,
        'U' => 30,
        'I' => 31,
        'O' => 32,
        'P' => 33,
        '{' => 34,
        '}' => 35,
        '|' => 51,

        // Row 3
        'A' => 38,
        'S' => 39,
        'D' => 40,
        'F' => 41,
        'G' => 42,
        'H' => 43,
        'J' => 44,
        'K' => 45,
        'L' => 46,
        ':' => 47,
        '"' => 48,

        // Row 4
        'Z' => 52,
        'X' => 53,
        'C' => 54,
        'V' => 55,
        'B' => 56,
        'N' => 57,
        'M' => 58,
        '<' => 59,
        '>' => 60,
        '?' => 61,
        _ => 0,
    }
}

/// Returns zero (0) if not matched.
pub fn get_lower_key(key: char) -> u8 {
     match key {
        // Row 1
        '1' => 10,
        '2' => 11,
        '3' => 12,
        '4' => 13,
        '5' => 14,
        '6' => 15,
        '7' => 16,
        '8' => 17,
        '9' => 18,
        '0' => 19,
        '-' => 20,
        '=' => 21,

        // Row 2
        'q' => 24,
        'w' => 25,
        'e' => 26,
        'r' => 27,
        't' => 28,
        'y' => 29,
        'u' => 30,
        'i' => 31,
        'o' => 32,
        'p' => 33,
        '[' => 34,
        ']' => 35,
        '\\' => 51,

        // Row 3
        'a' => 38,
        's' => 39,
        'd' => 40,
        'f' => 41,
        'g' => 42,
        'h' => 43,
        'j' => 44,
        'k' => 45,
        'l' => 46,
        ';' => 47,
        '\'' => 48,

        // Row 4
        'z' => 52,
        'x' => 53,
        'c' => 54,
        'v' => 55,
        'b' => 56,
        'n' => 57,
        'm' => 58,
        ',' => 59,
        '.' => 60,
        '/' => 61,
        _ => 0,
    }
}

/// Gets a keycode, and true if it is "upper".
pub fn get_keycode(key: char) -> Option<(u8, bool)> {
    match get_lower_key(key) {
        0 => {
            match get_upper_key(key) {
                0 => None,
                v => Some((v, true)),
            }
        },
        v => Some((v, false))
    }
}

#[cfg(test)]
mod tests {

    const LOWERS: &str = "abcdefghijklmnopqrstuvwxyz0123456789[]\\';/.,";
    const UPPERS: &str = "ABCDEFGHIJKLMNOPQRSTUV!@#$%^&*()_+{}|\":?><";
    const SPECIALS: [&str; 7] = ["space", "return", "tab", "up", "down", "left", "right"];

    #[test]
    fn test_get_lower_key() {
        // All lowers should return
        for c in LOWERS.chars() {
            let lower = super::get_lower_key(c);
            assert!(lower != 0);
        }
        // No uppers should return
        for c in UPPERS.chars() {
            let upper = super::get_lower_key(c);
            assert!(upper == 0);
        }
    }

    #[test]
    fn test_get_upper_key() {
        // All uppers should return
        for c in UPPERS.chars() {
            let upper = super::get_upper_key(c);
            assert!(upper != 0);
        }
        // No lowers should return
        for c in LOWERS.chars() {
            let lower = super::get_upper_key(c);
            assert!(lower == 0);
        }
    }

    #[test]
    fn test_get_special() {
        // All specials should return
        for s in SPECIALS.iter() {
            let code = super::special_from_str(s);
            assert!(!code.is_err());
        }
        assert!(super::special_from_str("garbage").is_err());
    }
}