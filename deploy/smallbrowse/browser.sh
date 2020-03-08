# Runs chromium in firejail, and restarts it if it dies.
# If a user closes the browser, it should just pop back up.

until firejail --profile=jail.conf --dns=1.1.1.1 --dns=8.8.4.4 --private chromium --no-remote --display=:10 --start-maximized --incognito duckduckgo.com
do
	echo "Chromium died: $?" >&2
	sleep 3
done
