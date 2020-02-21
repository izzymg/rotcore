#include <stdlib.h>
#include <stdio.h>
#include <zmq.h>
#include <czmq.h>
#include <signal.h>
#include <xdo.h>
#include <assert.h>
#include <math.h>
#include <glib.h>

/* XInteract: X11 m/kb interaction via zeromq. izzymg. */

#define SCREEN_WIDTH 1280
#define SCREEN_HEIGHT 720
#define MAX_EVENT_LEN 20
#define POLL_FREQ_MS 20
#define NUM_SPECIALS 8
#define SPECIALS_LEN 10

static int run = 1;
// Sorted list of allowed special characters
static char allowed_specials[NUM_SPECIALS][SPECIALS_LEN] = { "BackSpace", "Down", "Left", "Return", "Right", "space", "Tab", "Up" };

// Sends the text string to the active window.
int xi_send_text_string(xdo_t *instance, const char *text) {
    return xdo_enter_text_window(
        instance, CURRENTWINDOW,
        text, 0
    );
}

// Sends a special sequence to the active window.
int xi_send_special(xdo_t *instance, char *special) {

    // Search for allowed special characters
    char *sequence = bsearch(
        special, allowed_specials,
        NUM_SPECIALS, SPECIALS_LEN,
        (int(*) (const void*, const void*)) strcmp
    );

    if(sequence == NULL) {
        g_message("Ignoring invalid sequence %s\n", special);
        return -1;
    }
    return xdo_send_keysequence_window(
        instance, CURRENTWINDOW,
        sequence, 0
    );
}

// Basic linear interpolate
int approach(int target, int current, int delta) {
    int diff = target - current;
    if(diff > delta) {
        return current + delta;
    }
    if(diff < -delta) {
        return current - delta;
    }
    return target;
}

// Scroll the window, down if y < 1.
int xi_scroll(xdo_t *instance, int y) {
    // 4 = up 5 = down
    if(y < 1) {
        y = 5;
    } else {
        y = 4;
    }

    return xdo_click_window(instance, CURRENTWINDOW, y);
}

/* Move the mouse closer by a delta to screen target screen percentages X & Y.
The caller should be repeatedly calling this to interpolate to the target.
*/
int xi_mouse_approach(xdo_t *instance, float percent_x, float percent_y) {
    // Clamp 0-1
    if(percent_x > 1.0f) percent_x = 1.0f;
    if(percent_x < 0.0f) percent_x = 0.0f;
    if(percent_y > 1.0f) percent_y = 1.0f;
    if(percent_y < 0.0f) percent_y = 0.0f;

    int target_x = percent_x * SCREEN_WIDTH;
    int target_y = percent_y * SCREEN_HEIGHT;

    // Interpolate by delta
    int current_x;
    int current_y;
    xdo_get_mouse_location(instance, &current_x, &current_y, 0);

    int new_x = approach(target_x, current_x, 10);
    int new_y = approach(target_y, current_y, 10);

    return xdo_move_mouse(instance, new_x, new_y, 0);
}


// Send a click event: 1 left 2 mid 3 right.
int xi_mouse_click(xdo_t *instance, int button) {
    if(button < 1) button = 1;
    if(button > 3) button = 3;
    return xdo_click_window(instance, CURRENTWINDOW, button);
}

void xi_stop_running(int sig) {
    run = 0;
}

// XInteract entry
int main(int argc, char **argv) {
    g_message("XInteract | xdo version %s\n", xdo_version());
    int major, minor, patch;
    zmq_version (&major, &minor, &patch);
    g_message("Using ZMQ version %d.%d.%d\n", major, minor, patch);

    // Initialize XDO
    xdo_t *xdo_instance = xdo_new(NULL);
    if(xdo_instance == NULL) {
        g_message("Xdo init failure, exiting\n");
        return EXIT_FAILURE;
    }

    zsock_t *sink = zsock_new(ZMQ_PULL);
    char *address = getenv("XI_ADDRESS");
    if(address == NULL) {
        if(argv[1] != NULL) {
            address = argv[1];
        } else {
            address = "tcp://127.0.0.1:9674";
        }
    }
    g_message("Using address %s\n", address);
    zsock_connect(sink, address);

    // Catch sigint. Done in an odd place as it often conflicts with czmq.
    signal(SIGINT, xi_stop_running);

    // Percentages of mouse x, y coordinates
    float mouse_x = 0.5;
    float mouse_y = 0.5;
    char *mouse_substr;

    int register_n = 0;

    char recv_event = '-';

    zpoller_t *poller = zpoller_new(sink, NULL);

    while(run) {
        // Poll the second every millisecond
        zsock_t *sock = zpoller_wait(poller, 5);

        if(sock == NULL) {
            /* Constantly interpolate mouse position
            to target even if there's no data. */
            xi_mouse_approach(xdo_instance, mouse_x, mouse_y);
            continue;
        }

        // Got data, pull string in
        char *data = zstr_recv(sock);
        if(!data) {
            g_message("Skip data\n");
            continue;
        }
        int datalen = strlen(data);
        if(datalen < 2 || datalen > MAX_EVENT_LEN) {
            g_message("Discarding, invalid data length\n");
            zstr_free(&data);
            continue;
        }

        recv_event = data[0];
        switch(recv_event) {
            // Move mouse, by format M .percx .percy
            case 'M': {
                if(strlen(data) > 10) {
                    break;
                }
                // Take first two floating point values
                mouse_x = strtof(data + 1, &mouse_substr);
                mouse_y = strtof(mouse_substr, NULL);
                break;
            }
            // Scroll mouse, by format S dir
            case 'S': {
                if(strlen(data) > 3) {
                    break;
                }
                register_n = atoi(data + 1);
                xi_scroll(xdo_instance, register_n);
                break;
            }
            // Click mouse, by format C button
            case 'C': {
                if(strlen(data) > 3) {
                    break;
                }
                register_n = atoi(data + 1);
                xi_mouse_click(xdo_instance, register_n);
                break;
            }
            case 'T': {
                if(strlen(data) > 20) {
                    break;
                }
                g_message("Text: %s\n", data);
                xi_send_text_string(xdo_instance, data + 2);
            }
            // Special character, by format X type
            case 'X': {
                if(strlen(data) > SPECIALS_LEN) {
                    break;
                }
                g_message("Special: %s\n", data);
                xi_send_special(xdo_instance, data + 2);
                break;
            }
            default:
                g_message("Ignoring unknown event\n");
                break;
        }
        recv_event = '-';
        zstr_free(&data);
    }

    g_message("XInteract exiting\n");
    xdo_free(xdo_instance);
    zpoller_destroy(&poller);
    zsock_destroy(&sink);
}