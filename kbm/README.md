### KBM

KBM is a Rust program which allows interaction with the mouse & keyboard in x11. It connects to the x11 `$DISPLAY`,
and starts a TCP server which authenticates only one single TCP stream if it provides an acceptable challenge.

That server responds to special commands to click, scroll, type, etc.

As such, it will want the same secret as RTC.