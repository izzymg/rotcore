// Example configuration for deployments.
// You should replace this file with `config.js`, which is ignored by git.
module.exports = {

  /**
  When set to false, chromium runs without a sandbox,
  which is necessary when inside a container (e.g. docker).
  Leave **true** unless absolutely necessary.
  */
 sandbox: true,

  /** The X display to use. Should be in the format of ":2",
  where 2 is the display integer. Include the quotes. */
  display: ":10",

  /** Address the WebRTC app binds its TCP listener on. */
  rtcAddress: "0.0.0.0:8080",
  /** Address the KBM app binds its TCP listener on. */
  kbmAddress: "0.0.0.0:8081",
};