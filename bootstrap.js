/*
Bootstrap begins all RotCore room required processes.
*/

const { spawn } = require("child_process");
const { readFileSync } = require("fs");
const path =  require("path");

const secretPath = path.join(__dirname, "secret.txt");

/**
 * Read the GStreamer pipeline into memory.
*/
const readPipeline = () => {
    console.log("Reading pipeline...");
    const pipeline = readFileSync("conf/send.pipeline", { encoding: "utf8" });
    return pipeline;
};

// x11, chromium in firejail, kbm, streamer, & rotcore
let processes = [
    {
        program: "X",
        args: ["-quiet", "-config", "conf/10-headless.conf", ":10"],
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
        args: ["0.0.0.0:9232", secretPath],
        env: { "DISPLAY": ":10", },
    },
    {
        program: "bin/rotcore",
        args: [`-secret=${secretPath}`],
    }
];
let running = [];

const exit = signal => {
    console.log(`Received ${signal}, cleaning up children`);
    running.forEach(child => child.kill(signal));
};

process.on("SIGINT", exit);
process.on("SIGTERM", exit);
process.on("beforeExit", exit);

(async () => {
    console.log("Bootstrap spawning children");
    for(const p of processes) {

        console.log(`Starting ${p.program}`);

        // Mix process.env with env configuration
        const child = spawn(p.program, p.args, { stdio: 'inherit', env: { ...process.env, ...p.env, } });

        child.on("error", err => {
            console.log(`${p.program} exited with err: ${err}`);
            exit("SIGKILL");
            return;
        });

        // Hook events
        child.on("exit", code => {
            console.log(`${p.program} exited with code: ${code}`);
            exit("SIGKILL");
            return;
        });
 
        running = [child, ...running];
        // Block, allow process time if needed
        if(p.wait) {
            await new Promise(res => setTimeout(res, p.wait));
        }
    }
})();