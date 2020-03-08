#!/bin/bash

# Boots Xorg & RatPoison WM

Xorg :10 -config 10-headless.conf &
sleep 2
DISPLAY=:10 ratpoison