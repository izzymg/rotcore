pub mod key;

use x11rb::errors::{ConnectionError};
use x11rb::xcb_ffi::{XCBConnection};
use x11rb::connection::{Connection};
use x11rb::generated::xproto;
use x11rb::generated::xtest;
use x11rb::cookie::{VoidCookie};

use std::sync::mpsc;
use std::thread;
use std::time;

/// Amount to jump the pixel by each movement during interpolation.
const MOUSE_DELTA: i16 = 1;

/// Time in ms a keypress blocks, avoiding repeating key bug.
const KEY_TIME: u64 = 50;

const SCREEN_WIDTH: f32 = 1280.0;
const SCREEN_HEIGHT: f32 = 720.0;

/// Clamps x between min & max.
fn clamp(x: i16, min: i16, max: i16) -> i16 {
    if x > max {
        max
    } else if x < min {
        min
    } else {
        x
    }
}

/// Interpolates current towards target by MOUSE_DELTA.
fn approach(target: i16, current: i16) -> i16 {
    let diff = target - current;
    if diff > MOUSE_DELTA {
        return current + MOUSE_DELTA;
    }
    if diff < (-MOUSE_DELTA) {
        return current - MOUSE_DELTA;
    }
    target
}

/// Gets the root ID of the active X screen.
fn get_connection_root(conn: &XCBConnection, screen_number: usize) -> u32 {
    let setup = &conn.setup();
    let root = setup.roots[screen_number].root;
    return root;
}

/// Represents a client connection to the x11 display.
pub struct XConnection {
    connection: XCBConnection,
    screen_root: u32,
}

impl XConnection {

    /// Creates a new XConnection, using the $DISPLAY environment variable.
    pub fn new() -> Result<XConnection, ConnectionError> {
        let (conn, screen_number) = x11rb::xcb_ffi::XCBConnection::connect(None)?;
        let root = get_connection_root(&conn, screen_number);

        let xx = XConnection {
            connection: conn,
            screen_root: root,
        };
        Ok(xx)
    }

    /// Teleports the mouse pointer to a location as a percentage from 0-100 on the screen.
    pub fn mouse_to(&self, x: i16, y: i16) {
        let x = clamp(x, 0, 100);
        let y = clamp(y, 0, 100);

        let px = SCREEN_WIDTH * (x as f32 / 100.0);
        let py = SCREEN_HEIGHT * (y as f32 / 100.0);

        xproto::warp_pointer(
            &self.connection,
            x11rb::NONE,
            self.screen_root,
            0, 0,
            0, 0,
            px as i16, py as i16
        ).unwrap();
        self.connection.flush();
    }

    // Interpolate mouse to x, y from c_x, c_y, returning new the position.
    pub fn mouse_approach(&self, target: (i16, i16), current: (i16, i16)) -> (i16, i16) {
        let (x, y) = (approach(target.0, current.0), approach(target.1, current.1));
        self.mouse_to(x, y);
        (x, y)
    }

    /// Blocks indefinitely, receiving x & y coordinates from the rx.
    pub fn mouse_loop(&self, rx: mpsc::Receiver<(i16, i16)>) {
        let mut current: (i16, i16) = (0, 0);
        let mut target: (i16, i16) = (0, 0);
        loop {
                // Return new target if received, otherwise move towards the same target.
                target = match rx.try_recv() {
                    Ok(new) => new,
                    Err(err) => match err {
                        mpsc::TryRecvError::Empty => target,
                        mpsc::TryRecvError::Disconnected => {
                            println!("Mouse loop channel disconnected");
                            return;
                        },
                    },
                };
                let new_current = self.mouse_approach(target, current);
                current = new_current;
        }
    }

    /// Raw key press/release.
    fn trigger_key(&self, keycode: u8, ev: key::InputEvent) -> Result<VoidCookie<XCBConnection>, ConnectionError> {
        let r = xtest::fake_input(&self.connection, ev as u8, keycode, 0, self.screen_root, 0, 0, 0);
        thread::sleep(time::Duration::from_millis(KEY_TIME));
        r
    }

    /// Toggles the upper-key state on.
    fn set_upper(&self) {
        xtest::fake_input(
            &self.connection,
            key::InputEvent::KeyPress as u8,
            key::SHIFT_KEYCODE,
            0,
            self.screen_root,
            0,
            0,
            0,
        ).unwrap();
    }

    /// Toggles the upper-key state off.
    fn release_upper(&self) {
        xtest::fake_input(
            &self.connection,
            key::InputEvent::KeyRelease as u8,
            key::SHIFT_KEYCODE,
            0,
            self.screen_root,
            0,
            0,
            0,
        ).unwrap();
    }

    /// Loops up a special character and presses it.
    fn enter_special(&self, code: key::SpecialCode) -> bool {
        let code = code as u8;
        self.trigger_key(code, key::InputEvent::KeyPress).unwrap();
        self.trigger_key(code, key::InputEvent::KeyRelease).unwrap();
        self.connection.flush();
        true
    }

    /// Looks up & presses a key by character, simulating a press & release.
    fn enter_character(&self, key: char) -> bool {
        let (code, is_upper) = match key::get_keycode(key) {
            Some((c, iu)) => (c, iu),
            None => return false,
        };

        if is_upper {
            self.set_upper();
        }

        self.trigger_key(code, key::InputEvent::KeyPress).unwrap();
        self.trigger_key(code, key::InputEvent::KeyRelease).unwrap();
        self.connection.flush();

        if is_upper {
            self.release_upper();
            self.connection.flush();
        }

        true
    }

    pub fn tap_mouse_button(&self, mc: key::MouseCode) {
        let keycode = mc as u8;
        self.trigger_key(keycode, key::InputEvent::MousePress).unwrap();
        self.trigger_key(keycode, key::InputEvent::MouseRelease).unwrap();
        self.connection.flush();
    }

    /// Blocks indefinitely, receiving keyboard button & mouse button inputs.
    pub fn button_loop(
        &self,
        key_inputs: mpsc::Receiver<char>,
        mouse_inputs: mpsc::Receiver<key::MouseCode>,
        special_inputs: mpsc::Receiver<key::SpecialCode>) {
        loop {
            match key_inputs.try_recv() {
                Ok(character) => { self.enter_character(character); },
                _ => (),
            }

            match mouse_inputs.try_recv() {
                Ok(button) => { self.tap_mouse_button(button); },
                _ => (),
            }

            match special_inputs.try_recv() {
                Ok(s) => { self.enter_special(s); },
                _ => (),
            }
        }
    }


}