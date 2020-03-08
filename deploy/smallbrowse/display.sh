# Starts a headless Xorg display on :10 using ratpoison WM
xinit /usr/bin/ratpoison -- :10 -xf86config 10-headless.conf -quiet