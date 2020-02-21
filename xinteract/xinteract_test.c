#include <stdlib.h>
#include <stdio.h>
#include <zmq.h>
#include <czmq.h>
#include <signal.h>
#include <xdo.h>
#include <assert.h>

/* XInteract: X11 m/kb interaction via zeromq. izzymg 2020 */

// XInteract entry
int main(int argc, char **argv) {
    int major, minor, patch;
    zmq_version (&major, &minor, &patch);
    printf("Using ZMQ version %d.%d.%d\n", major, minor, patch);

    zsock_t *publisher = zsock_new(ZMQ_PUSH);
    zsock_bind(publisher, "tcp://127.0.0.1:9674");

    for(int i = 0; i < 3; i++) {
        zstr_send(publisher, "M .98 .02");
        zstr_send(publisher, "M .98 .02");
        zstr_send(publisher        , "M .01 .65");
        zstr_send(publisher, "M .88 .89");
        zstr_send(publisher, "M ankwjd o2i>");
        zstr_send(publisher, "M ankwjd o2i>");
        zstr_send(publisher, "S -1");
        zstr_send(publisher, "S 300");
        zstr_send(publisher,        zstr_send(publisher, "S 1");
"S 0");
        zstr_send(publisher, "S 02");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "S asmdlaslkmdl");
        zstr_send(publisher, "C 1");
        zstr_send(publisher, "C 2");
        zstr_send(publisher, "S 1");
        zstr_send(publisher, "C 3");
        zstr_send(publisher, "C 12309132909");
        sleep(2);
    }

    printf("XInteract exiting\n");
    zsock_destroy(&publisher);
}