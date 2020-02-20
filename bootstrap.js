/*
Bootstrap begins all RotCore room required processes,
including an X11 display, jailed chromium process, WebRTC SFU,
XSend & XInteract. IzzyMG.
*/

const { spawn } = require("child_process");

let processes = [
    {
        program: "X",
        args: ["-config", "conf/10-headless.conf", ":10"],
    },
];


/**
 * Simple console log wrapper.
 * @param {boolean} err Is this an error log, or stdout.
 * @param  {...any} data Data to be logged.
*/
const log = (err, ...data) => {
    let write = console.log;
    if(err) {
        write = console.error;
    }
    write(new Date(Date.now()).toLocaleString(), ...data);
};

// Spawn each process and hook into its stdout/stderr/close
log(false, "Spawning children");
let running = [];
processes.forEach(process => {
    const child = spawn(process.program, process.args);
    child.stdout.on("data", data => log(false, `${process.program}: ${data}`));
    child.stderr.on("data", data => log(true, `${process.program} Error: ${data}`));
    child.on("exit", code => log(true, `${process.program} exited with code: ${code}`));

    running = [child, ...running];
});

const exit = signal => {
    log(false, `Received ${signal}, cleaning up children`);
    running.forEach(child => child.kill(signal));
};

process.on("SIGINT", exit);
process.on("SIGTERM", exit);