/*
Bootstrap begins all RotCore room required processes.
*/

const { spawn } = require("child_process");
const { readFileSync } = require("fs");
const path =  require("path");

const secretPath = path.join(__dirname, "conf/secret");

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


/**
 * Read the GStreamer pipeline into memory.
*/
const readPipeline = () => {
    log(false, "Reading pipeline...");
    const pipeline = readFileSync("conf/send.pipeline", { encoding: "utf8" });
    log(false, pipeline);
    return pipeline;
};

// x11, chromium in firejail, kbm, streamer, & rotcore
let processes = [
    {
        program: "X",
        args: ["-config", "conf/10-headless.conf", ":10"],
        wait: 2000,
    },
    {
        program: "firejail",
        args: ["--profile=conf/jail.conf","--private", "--dns=1.1.1.1", "--dns=8.8.4.4", "chromium", "--no-remote"],
        env: { "DISPLAY": ":10" },
    },
    {
        program: "bin/streamer",
        args: [readPipeline()],
        env: { "DISPLAY": ":10" },
    },
    {
        program: "bin/kbm/release/kbm",
        args: ["127.0.0.1:9232"],
        env: { "DISPLAY": ":10", },
    },
    {
        program: "bin/rotcore",
    }
];
let running = [];

const exit = signal => {
    console.log("Exiting");
    log(false, `Received ${signal}, cleaning up children`);
    running.forEach(child => child.kill(signal));
    setTimeout(() => process.exit(0), 500);
};

process.on("SIGINT", exit);
process.on("SIGTERM", exit);
process.on("beforeExit", exit);

(async () => {
    log(false, "Bootstrap spawning children");
    for(const p of processes) {

        log(false, `Starting ${p.program}`);

        // Mix process.env with env configuration
        const child = spawn(p.program, p.args, { stdio: 'inherit', env: { ...process.env, ...p.env, } });

        // Hook events
        child.on("exit", code => {
            log(true, `${p.program} exited with code: ${code}`);
            process.exit(1);
        });
 
        running = [child, ...running];
        // Block, allow process time if needed
        if(p.wait) {
            await new Promise(res => setTimeout(res, p.wait));
        }
    }
})();