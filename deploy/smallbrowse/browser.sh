# Runs chromium in firejail, and restarts it if it dies.
# If a user closes the browser, it should just pop back up.

while [ 1 ]
do
        if firejail --private --seccomp --profile=jail.conf --dns=1.1.1.1 --dns=8.8.4.4 chromium --no-remote --display=:10 --start-maximized --incognito duckduckgo.com; then
                echo "Chromium died: $?" >&2
                sleep 1
        else
                echo "Chromium crashed?: $?" >&2
                sleep 1
        fi
done
