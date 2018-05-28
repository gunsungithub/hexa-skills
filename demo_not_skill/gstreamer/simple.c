#include <gst/gst.h>
#include <glib.h>


static gboolean bus_call (GstBus *bus, GstMessage *msg, gpointer data)
{
    GMainLoop *loop = (GMainLoop *) data;

    switch (GST_MESSAGE_TYPE (msg)) {
        case GST_MESSAGE_EOS:
            g_print ("End of stream\n");
            g_main_loop_quit (loop);
            break;

        case GST_MESSAGE_ERROR: {
            gchar  *debug;
            GError *error;

            gst_message_parse_error (msg, &error, &debug);
            g_free (debug);

            g_printerr ("Error: %s\n", error->message);
            g_error_free (error);

            g_main_loop_quit (loop);
            break;
        }
        default:
            break;
    }

    return TRUE;
}

int main (int argc,char *argv[])
{
    GMainLoop *loop;

    GstElement *pipeline, *videosrc, *text, *videoenc, *videoconvert, *muxer, *sink;
    GstBus *bus;
    guint bus_watch_id;

    if (argc < 2) {
        g_printerr ("Usage: %s <rtmp url>\n", argv[0]);
        return -1;
    }
    /* Initialisation */
    gst_init (&argc, &argv);

    loop = g_main_loop_new (NULL, FALSE);

    /* Create gstreamer elements */
    pipeline = gst_pipeline_new ("media-player");
    videosrc = gst_element_factory_make ("v4l2src",         "video-camrta-source");
    text     = gst_element_factory_make ("textoverlay",         "text");
    videoenc = gst_element_factory_make ("imxvpuenc_h264",  "video-h264-byte-stream");
    videoconvert = gst_element_factory_make ("h264parse",       "video-convert");
    muxer    = gst_element_factory_make ("flvmux",          "flv-muxer");
    sink     = gst_element_factory_make ("rtmpsink",      "sink");

    if (!pipeline || !videosrc || !text || !videoenc || !videoconvert || !muxer || !sink) {
        g_printerr ("One element could not be created. Exiting.\n");
        return -1;
    }

    /* Set up the pipeline */
    /* we set the input filename to the source element */
    g_object_set (G_OBJECT (sink), "location", argv[1], NULL);
    g_object_set (G_OBJECT (text), "text", argv[1], NULL);
    /* we add a message handler */
    bus = gst_pipeline_get_bus (GST_PIPELINE (pipeline));
    bus_watch_id = gst_bus_add_watch (bus, bus_call, loop);
    gst_object_unref (bus);

    /* we add all elements into the pipeline */
    gst_bin_add_many (GST_BIN (pipeline), videosrc, text, videoenc, videoconvert, muxer, sink, NULL);
    /* we link the elements together */
    if (gst_element_link (videosrc, text)){
        g_print ("link success %d\n", __LINE__);
    }
    else{
        return -1;
    }       
    if (gst_element_link (text, videoenc)){
        g_print ("link success %d\n", __LINE__);
    }
    else{
        return -1;
    }
    if (gst_element_link (videoenc, videoconvert)){
        g_print ("link success %d\n", __LINE__);
    }
    else{
        return -1;
    }
    if (gst_element_link (videoconvert, muxer)){
        g_print ("link success %d\n", __LINE__);
    }
    else{
        return -1;
    }
    if (gst_element_link (muxer, sink)){
        g_print ("link success %d\n", __LINE__);
    }
    else{
        return -1;
    }

    /* Set the pipeline to "playing" state*/
    //g_print ("Now playing: %s\n", argv[1]);
    gst_element_set_state (pipeline, GST_STATE_PLAYING);

    /* Iterate */
    g_print ("Running...\n");
    g_main_loop_run (loop);

    /* Out of the main loop, clean up nicely */
    g_print ("Returned, stopping playback\n");
    gst_element_set_state (pipeline, GST_STATE_NULL);

    g_print ("Deleting pipeline\n");
    gst_object_unref (GST_OBJECT (pipeline));
    g_source_remove (bus_watch_id);
    g_main_loop_unref (loop);
    return 0;
}
