/* Bootstraps all processes required to run an RTC room:

1: x11 display
2: browser
3: stream.sh, to start gstreamer UDP streaming of the desktop & audio 
4: rotcore, to take stream data and respond to WebRTC requests
5: KBM, to allow for keyboard & mouse interaction over TCP

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
    this.stop = () => {
        if(!child) {
            return;
        }
        console.log(`${this.exe} exiting.`);
        child.kill("SIGTERM");
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
    const x11 = new Process({
        directory: "smallbrowse",
        executable: "display.sh",
    });

    const browser = new Process({
        directory: "smallbrowse",
        executable: "browser.sh",
    });

    // Stream and KBM need to know the X11 display in use.

    const stream = new Process({
        directory: ".",
        executable: "stream.sh",
        environment: { DISPLAY: config.display, }
    });

    const kbm = new Process({
        directory: "bin/kbm/release",
        executable: "kbm",
        environment: { DISPLAY: config.display, },
        args: [config.kbmAddress, secretFilePath],
    });

    const rotcore = new Process({
        directory: "bin",
        executable: "rotcore",
        args: [...config.publicIps.map(ip => `--ip=${ip}`), `--secret=${secretFilePath}`],
        environment: {
            SIGNAL_ADDRESS: config.rtcAddress,
        }
    });

    // Setup clean quit when the parent dies.

    const exit = function() {
        console.log("Quitting children.");
        kbm.stop();
        stream.stop();
        x11.stop();
        rotcore.stop();
        browser.stop();
    };

    process.on("SIGINT", exit);
    process.on("SIGTERM", exit);

    x11.start();
    browser.start();

    // Give x11 time to come alive before starting the rest.
    setTimeout(() => {
        kbm.start();
        rotcore.start();
        stream.start();
    }, 500);
}

init();
