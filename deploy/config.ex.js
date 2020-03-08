// Example configuration for deployments.
// You should replace this file with `config.js`, which is ignored by git.

const config = {
  /** A string-array of all public IPs that should
  be sent to peers, for DNAT. */
  publicIps: [
    "203.111.33.30",
  ],

  /**
   * The X display to use. Should be in the format of ":2",
  where 2 is the display integer. Include the quotes. */
  display: ":10",

  /** Address the WebRTC app binds its TCP listener on. */
  rtcAddress: "0.0.0.0:8080",
  /** Address the KBM app binds its TCP listener on. */
  kbmAddress: "0.0.0.0:9232",
};

module.exports = config;