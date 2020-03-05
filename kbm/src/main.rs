/*
Simple x11 server for RotCore, starts a TPC server
and responds to requests for keyboard & mouse input
using x11rb.
*/

mod x11;
mod auth;

use std::io::prelude::*;
use std::net;
use std::io;
use std::time;
use std::str;
use std::env;
use std::fs;

use std::sync::Arc;
use std::sync::mpsc;

use std::thread;


/// Maximum incoming payload in bytes that can be read.
const MAX_INCOMING_READ: usize = 100;

const MOUSE: &[u8] = &[b'c'];
const TEXT: &[u8] = &[b't'];
const POINTER: &[u8] = &[b'm'];
const SPECIAL: &[u8] = &[b's'];

type PointerSender = mpsc::Sender<(i16, i16)>;
type KBSender = mpsc::Sender<char>;
type MBSender = mpsc::Sender<x11::key::MouseCode>;
type SSender = mpsc::Sender<x11::key::SpecialCode>;

/// Read from the stream, returning the number of bytes read & a buffer, or an error.
fn tcp_stream_read(mut stream: &net::TcpStream) -> Result<(usize, [u8; MAX_INCOMING_READ]), io::Error> {
    let mut buffer = [0; MAX_INCOMING_READ];

    match stream.read(&mut buffer) {
        Ok(bytes_read) => Ok((bytes_read, buffer)),
        Err(err) => Err(err),
    }
}


/// Parse a slice of bytes as a utf-8 numerical integer.
fn bytes_as_numerical<T>(x: &[u8]) -> Option::<T>
where T: std::str::FromStr {
    let x = match str::from_utf8(x) {
        Ok(x) => x,
        Err(e) => {
            println!("Failed to interpret utf8 bytes: {}", e);
            return None;
        },
    };

    let x = match x.parse::<T>() {
        Ok(x) => x,
        Err(_) => {
            println!("Failed to parse bytes as numerical");
            return None;
        },
    };
    Some(x)
}

struct Server {
    pub pointer_sender: PointerSender,
    pub kb_sender: KBSender,
    pub mb_sender: MBSender,
    pub s_sender: SSender,
} 

impl Server {

    /// Start listening on address. Lazy evals cmdline args for address.
    fn listen(&self) {

        let args: Vec<String> = env::args().collect();

        // Grab server address from args
        let mut iter = args.iter().skip(1);
        let address = match iter.next() {
            Some(a) => &a[0..a.len()],
            None => "127.0.0.1:8080",
        };

        // Grab server key from args
        let key = match iter.next() {
            Some(v) => &v[0..v.len()],
            None => "./secret.txt",
        };
        let key = fs::read_to_string(key)
            .expect("Failed to read secret file");
        let key = key.trim()
            .trim_right_matches("\r\n")
            .trim_right_matches('\n')
            .as_bytes();
        println!("{:?}", key);
        let key = auth::make_key(key);

        println!("Spawning server on {}", address);
        let listener = net::TcpListener::bind(address).unwrap();

        for stream in listener.incoming() {
            match stream {
                Ok(stream) => {
                    println!("Connection established~");
                    self.authorize_conn(stream, &key);
                },
                Err(err) => println!("Connection failure: {}", err),
            }
            println!("Finished connection");
        }
    }

    /// Check that a connection is authorized before processing its stream.
    fn authorize_conn(&self, stream: net::TcpStream, key: &auth::Key) {

        // Read the next stream message
        stream.set_read_timeout(Some(time::Duration::from_secs(5))).unwrap();
        stream.set_write_timeout(Some(time::Duration::from_secs(5))).unwrap();
        let read = tcp_stream_read(&stream);

        let (bytes_read, buf) = match read {
            Ok(v) => v,
            Err(err) => { println!("Read timeout: {}", err); return; },
        };

        // Split into "{data} {MAC}"
        let mut iter = buf[0..bytes_read].split(|b| b.is_ascii_whitespace());
        let data = match iter.next() {
            Some(v) => v,
            None => { println!("Invalid auth, no data"); return; }
        };
        let mac = match iter.next() {
            Some(v) => v,
            None => { println!("Invalid auth, no MAC"); return; }
        };

        if data.len() < 10 {
            println!("Invalid auth, data too small");
            return;
        }

        // Verify the data is signed with the same key before accepting stream.
        match auth::verify_hash(&data, &mac, &key) {
            Ok(v) => {
                if v == true {
                    self.read_loop(stream)
                } else {
                    println!("Invalid auth");
                    return;
                }
            },
            Err(e) => {
                println!("Auth error: {}", e);
                return;
            }
        }
    }

    /// Core loop, owns a stream, dispatches messages.
    fn read_loop(&self, stream: net::TcpStream) {
        stream.set_read_timeout(None).unwrap();
        loop {
            match tcp_stream_read(&stream) {
                Ok((bytes_read, buf)) => {
                    if bytes_read > 0 {
                        self.handle_incoming(&buf[0..bytes_read]);
                    } else {
                        println!("Connection gone");
                        return;
                    }
                },
                Err(err) => {
                    println!("Read error: {}", err);
                    return;
                }
            }
        };
    }

    /// Handles incoming stream messages.
    fn handle_incoming(&self, buf: &[u8]) {
        let mut iter = buf.split(|b| b.is_ascii_whitespace());
        let result = match iter.next() {
            Some(ev) => match ev {
                TEXT => self.on_kb_message(iter),
                MOUSE => self.on_mb_message(iter),
                POINTER => self.on_pointer_message(iter),
                SPECIAL => self.on_special_message(iter),
                _ => None,
            },

            // Fallback
            _ => None,
        };

        match result {
            Some(x) => {
                if x == false {
                    println!("Operation did not succeed: error in event");
                }
            }
            None => {
                println!("Operation did not succeed: error in data");
            }
        }
    }

    fn on_special_message<'a, I>(&self, mut it: I) -> Option<bool>
    where I: Iterator<Item = &'a [u8]> {
        // Interpret next bytes as a string containing a special keypress
        let key = it.next()?;
        if key.len() < 1 {
            return None;
        }

        let key = match str::from_utf8(key) {
            Ok(s) => s,
            Err(e) => {
                println!("Parsing special failed: {}", e);
                return Some(false);
            }
        };

        let key = match x11::key::special_from_str(key) {
            Ok(k) => k,
            Err(e) => {
                println!("Parsing special failed: {}", e);
                return Some(false);
            }
        };

        self.s_sender.send(key).unwrap();
        Some(true)
    }

    fn on_kb_message<'a, I>(&self, mut it: I) -> Option<bool>
    where I: Iterator<Item = &'a [u8]> {
        // Interpret next bytes as a keypress
        let key = it.next()?;
        if key.len() > 0 {
            self.kb_sender.send(key[0] as char).unwrap();
            Some(true)
        } else {
            Some(false)
        }
    }

    fn on_mb_message<'a, I>(&self, mut it: I) -> Option<bool>
    where I: Iterator<Item = &'a [u8]> {

        // Fetch next bytes from message
        let button = it.next()?;
        if button.len() < 1 {
            return None;
        }
        // Interpret as a numerical string
        let button = match bytes_as_numerical::<u8>(button) {
            Some(b) => b,
            None => {
                return Some(false);
            },
        };

        // Parse as a mouse code and trigger it.
        let button = match x11::key::mousecode_from_u8(button) {
            Ok(b) => b,
            Err(e) => {
                println!("MB error: {}", e);
                return Some(false);
            },
        };

        self.mb_sender.send(button).unwrap();
        Some(true)
    }

    fn on_pointer_message<'a, I>(&self, mut it: I) -> Option<bool>
    where I: Iterator<Item = &'a [u8]> {

        // Take next two slices as x, y coords
        let x_bytes = it.next()?;
        let y_bytes = it.next()?;

        let x = match bytes_as_numerical::<i16>(x_bytes) {
            Some(x) => x,
            None => {
                return Some(false)
            }
        };

        let y = match bytes_as_numerical::<i16>(y_bytes) {
            Some(y) => y,
            None => {
                return Some(false)
            }
        };

        // Update new pointer coords

        self.pointer_sender.send((x, y)).unwrap();
        Some(true)
    }
}

/// Entry point.
fn main() {
    println!("Connecting to X11");

    let x = match x11::XConnection::new() {
        Ok(x) => Arc::new(x),
        Err(err) => {
            println!("Failed to establish X11: {}", err);
            return
        }
    };

    /*
    Start x11 listening for mouse input on a new thread.
    */
    let (mtx, mouse_rx): (
        mpsc::Sender<(i16, i16)>, mpsc::Receiver<(i16, i16)>
    ) = mpsc::channel();

    let xm = x.clone();
    thread::spawn(move || {
        xm.mouse_loop(mouse_rx)
    });

    /*
    Start x11 listening for keyboard & mouse buttons on a new thread.
    */
    let (kbtx, kbrx): (
        mpsc::Sender<char>, mpsc::Receiver<char>
    ) = mpsc::channel();

    let (mbtx, mbrx): (
        mpsc::Sender<x11::key::MouseCode>, mpsc::Receiver<x11::key::MouseCode>
    ) = mpsc::channel();

    let (stx, srx): (
        mpsc::Sender<x11::key::SpecialCode>, mpsc::Receiver<x11::key::SpecialCode>
    ) = mpsc::channel();


    let xb = x.clone();
    thread::spawn(move || {
        xb.button_loop(kbrx, mbrx, srx)
    });

    /*
    Spawn TCP server.
    */

    let server = Server{
        pointer_sender: mtx,
        kb_sender: kbtx,
        mb_sender: mbtx,
        s_sender: stx,
    };

    server.listen();
}