/* Bootstraps all processes required to run an RTC room:

1: stream.sh, to start gstreamer UDP streaming of the desktop & audio 
2: rotcore, to take stream data and respond to WebRTC requests
3: KBM, to allow for keyboard & mouse interaction over TCP

*/

const config = require("./config");
const { join: pathJoin } = require("path");
const { spawn } = require("child_process");

// Location of the text file containing the secret for apps to use.
const secretFilePath = pathJoin(__dirname, "secret.txt");

/**
    Constructs a new process object.
    Processes have their environment setting mixed in with `process.env`,
    to allow for external envs to affect the program. Otherwise important things like
    XDG_RUNTIME_DIR and so on will not be passed down to children.

    The processes are started in "directory" so they may use relative paths.
*/
const Process = function({ directory, executable, args, environment, }) {

    this.directory = directory || "./";
    this.exe = pathJoin(__dirname, this.directory, executable || "noop");
    this.args = args || [];
    this.env = { ...process.env, ...(environment || {  }) };

    let child;

    /** Quits the process if its active with a SIGTERM. */
    this.stop = kill => {
        if(!child) {
            return;
        }
        console.log(`${this.exe} exiting.`);
        child.kill(kill ? "SIGKILL" : "SIGTERM");
    };

    /** Spawn the process in the given work-dir. */
    this.start = () => {
        console.log(`Starting ${this.exe} ${this.args.join(" ")}`);
        child = spawn(this.exe, this.args, {
            stdio: "inherit",
            cwd: this.directory,
            env: this.env,
        });
    };
};

const init = function() {

    // Process the public IPs from the environment as a comma separated list.
    let ipArgs = [];
    if(process.env.PUBLIC_IPS && process.env.PUBLIC_IPS.length > 0) {
        ipArgs = process.env.PUBLIC_IPS.split(",").map(ip => `--ip=${ip}`);
    }

    // Stream and KBM need to know the X11 display in use.

    const display = new Process({
        directory: "./display",
        executable: "display.sh",
        environment: { DISPLAY: config.display, }
    });

    /* Whether the browser should be sandboxed.
    See: config.ex.js */
    let sandboxArg = "";
    if(config.sandbox === false) {
        sandboxArg = "--no-sandbox";
    }
    console.log("SANDBOX ARGS:", sandboxArg);
    const browse = new Process({
        directory: "./browse",
        executable: "browse.sh",
        environment: { DISPLAY: config.display },
        args: [sandboxArg],
    });

    const rotcore = new Process({
        directory: "./bin",
        executable: "rtc",
        args: [...ipArgs, `--secret=${secretFilePath}`],
        environment: {
            SIGNAL_ADDRESS: config.rtcAddress,
        }
    });

    // Setup clean quit when the parent dies.

    const exit = function() {
        console.log("Quitting children.");
        display.stop();
        rotcore.stop();
        browse.stop(true);
    };

    process.on("SIGINT", exit);
    process.on("SIGTERM", exit);

    display.start();
    // Give a bit for Xorg to start
    setTimeout(() => {
        browse.start();
        rotcore.start();
    }, 1000);
}

init();
