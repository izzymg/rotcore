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
        wait: 2000,
    },
    {
        program: "firejail",
        args: ["--profile=conf/jail.conf","--private", "--dns=1.1.1.1", "--dns=8.8.4.4", "chromium", "--no-remote"],
        env: { "DISPLAY": ":10" }
    }
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


(async () => {
    let running = [];
    log(false, "Bootstrap spawning children");
    for(const p of processes) {

        // Mix process.env with env configuration
        const child = spawn(p.program, p.args, { env: { ...process.env, ...p.env, } });

        // Hook events
        child.stdout.on("data", data => log(false, `${p.program}: ${data}`));
        child.stderr.on("data", data => log(true, `${p.program}: ${data}`));
        child.on("exit", code => log(true, `${p.program} exited with code: ${code}`));

        running = [child, ...running];
        // Block, allow process time if needed
        if(p.wait) {
            await new Promise(res => setTimeout(res, p.wait));
        }
    }

    const exit = signal => {
        log(false, `Received ${signal}, cleaning up children`);
        running.forEach(child => child.kill(signal));
    };
    
    process.on("SIGINT", exit);
    process.on("SIGTERM", exit);
})();