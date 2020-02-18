#include <stdlib.h>
#include <stdio.h>
#include <zmq.h>
#include <czmq.h>
#include <signal.h>
#include <xdo.h>
#include <assert.h>

/* XInteract: X11 m/kb interaction via zeromq. izzymg 2020 */

#define EVENT "X11 %c %d %d"
#define MAX_EVENT_LEN 20

static int run = 1;

// Sends the text string to the active window.
int xi_send_text_string(xdo_t *instance, const char* text) {
    return xdo_enter_text_window(
        instance, CURRENTWINDOW,
        text, 0
    );
}

// Sends a special sequence to the active window.
int xi_send_special(xdo_t *instance, const char* sequence) {
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

// Scroll the window in the direction of Y.
int xi_scroll(xdo_t *instance, int y) {
    int button = 4;
    if(y < 0) {
        button = 5;
    }
    return xdo_click_window(instance, CURRENTWINDOW, button);
}

/* Move the mouse closer by a delta to screen target coordinates X & Y.
The caller should be repeatedly calling this to interpolate to the target.
*/
int xi_mouse_approach(xdo_t *instance, int target_x, int target_y) {
    int current_x;
    int current_y;
    xdo_get_mouse_location(instance, &current_x, &current_y, 0);

    int new_x = approach(target_x, current_x, 2);
    int new_y = approach(target_y, current_y, 2);

    return xdo_move_mouse(instance, new_x, new_y, 0);
}


// Send a click event.
int xi_mouse_click(xdo_t *instance, int right_click) {
    int button = 1;
    if(right_click) {
        button = 3;
    }
    return xdo_click_window(instance, CURRENTWINDOW, button);
}

void xi_stop_running(int sig) {
    run = 0;
}

// XInteract entry
int main(int argc, char **argv) {
    printf("XInteract | xdo version %s\n", xdo_version());
    int major, minor, patch;
    zmq_version (&major, &minor, &patch);
    printf("Using ZMQ version %d.%d.%d\n", major, minor, patch);

    // Initialize XDO
    xdo_t *xdo_instance = xdo_new(NULL);
    if(xdo_instance == NULL) {
        printf("Xdo init failure, exiting\n");
        return EXIT_FAILURE;
    }

    /*
    Read certificate files from environment variables and 
    apply own certificate to self, as well as setting the server's
    public key.
    */
    zsock_t *subscriber = zsock_new(ZMQ_SUB);
    zsock_set_subscribe(subscriber, "X11");

    char *address = getenv("XI_ADDRESS");
    if(address == NULL) {
        if(argv[1] != NULL) {
            address = argv[1];
        } else {
            address = "tcp://localhost:9673";
        }
    }
    printf("Using address %s\n", address);

    char *user = getenv("XI_USER");
    char *password = getenv("XI_PASSWORD");
    if(user != NULL) {
        zsock_set_plain_username(subscriber, user);
        printf("Set username\n");
    }
    if(password != NULL) {
        zsock_set_plain_password(subscriber, password);
        printf("Set password\n");
    }

    zsock_connect(subscriber, address);

    // Catch sigint. Done in an odd place as it often conflicts with czmq.
    signal(SIGINT, xi_stop_running);

    int mouse_x = 1280/2;
    int mouse_y = 720/2;
    char recv_event = '-';
    int recv_i;
    int recv_j;

    while(run) {

        /* Constantly interpolate mouse position
        to target even if there's no data. */
        //xi_mouse_approach(xdo_instance, x, y);

        char *data = zstr_recv_nowait(subscriber);
        if(data != NULL) {
            if(strlen(data) > MAX_EVENT_LEN) {
                zstr_free(&data);
                continue;
            }
            printf("Data\n");

            sscanf(data, "X11 %c %d %d", &recv_event, &recv_i, &recv_j);
            switch(recv_event) {
                case 'M':
                    // X11 M x y
                    mouse_x = recv_i;
                    mouse_y = recv_j;
                case 'S':
                    // X11 S dir
                    xi_scroll(xdo_instance, recv_i);
                    break;
                case 'C':
                    // X11 C button
                    xi_mouse_click(xdo_instance, recv_i);
                    break;
                default:
                    printf("Ignoring unknown event");
                    break;
            }
            recv_event = '-';
            zstr_free(&data);
        }
    }

    printf("XInteract exiting\n");
    xdo_free(xdo_instance);
    zsock_destroy(&subscriber);
}