#include <gst/gst.h>

// XSend: run a gstreamer pipeline. Izzymg.

int main(int argc, char **argv) {
    if(argc < 2) {
        g_error("Expected 1 argument: pipeline");
        return EXIT_FAILURE;
    }
    gst_init(&argc, &argv);
    GError *err = NULL;
    GstElement *pipeline = gst_parse_launch(argv[1], &err);
    if(err != NULL) {
        g_error(err->message);
        g_error_free(err);
        return EXIT_FAILURE;
    }

    int ret = gst_element_set_state(pipeline, GST_STATE_PLAYING);
    g_message("%d RET\n", ret);
    g_assert(ret != GST_STATE_CHANGE_FAILURE);

    g_message("Pipeline setup\n");

    gst_deinit();
    gst_object_unref(pipeline);
    return 0;
}