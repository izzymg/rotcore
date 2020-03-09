# Runs a browser, and restarts it if it dies.

while [ 1 ]
do
        if firefox --display=$DISPLAY --private "https://youtu.be/-QezMBY0Ndk?t=15"; then
                echo "Browser died: $?" >&2
                sleep 3
        else
                echo "Browser crashed: $?" >&2
                sleep 3
        fi
done
