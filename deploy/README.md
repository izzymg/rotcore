# RotCore: Deploy

This directory contains a list of scripts to deploy Rotcore, including starting the streaming service,
RTC service, KBM service, Streamer service and a secure X11 environment with a browser.

## Environment

The most necessary thing to set is `PUBLIC_IPS`, which is parsed as a **comma separated list** for 1:1 DNAT mapping
of public IP addresses. These are used in the host section of outgoing SDP's in WebRTC negotiations.