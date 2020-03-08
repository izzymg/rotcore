# RotCore: Deploy

This directory contains a list of scripts to deploy Rotcore, including starting the streaming service,
WebRTC service, KBM service, and a secure X11 environment with a browser.

## Run

Run `make` one directory up (in the root of Rotcore) to generate all the binaries required.

Copy `config.ex.js` -> `config.js` and fill in any information required (see the comments documenting it).

Then just run `node deploy.js` to bootstrap the entire thing and run a room!

**Important**: Run `deploy.js`, from *within* the deploy folder. It relies on relative paths to the scripts.