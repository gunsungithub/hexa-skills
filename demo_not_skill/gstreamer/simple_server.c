#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include <string.h>
#include <fcntl.h>
#include <errno.h>
#include <sys/time.h>

#include <gst/gst.h>
#include <glib.h>

#define FIFO_FILE "/tmp/myfifo"

int init_daemon(void)
{
    int i;
    pid_t pid = fork();
    if (pid < 0){
        return -1;
    }
    else if (pid){
        exit(0);                       /* parent terminates */
    }
    if (setsid() < 0){
        return -1;
    }
    signal(SIGHUP, SIG_IGN);
    pid = fork();
    if (pid < 0){
        return -1;
    }
    else if (pid){
        exit(0);
    }
    chdir("/");
    /* close all fd */
    for (i = 0; i < 255; i++){
        close(i);
    }
    /* redirect stdin, stdout, stderr to /dev/null */
    open("/dev/null", O_RDONLY);
    open("/dev/null", O_RDWR);
    open("/dev/null", O_RDWR);
    return 0;
}

static GMainLoop *loop = NULL;
static GstElement *pipeline;
static void *start_loop(){
    g_print ("Running...\n");
    g_main_loop_run (loop);

    g_print ("Returned, stopping playback\n");
    gst_element_set_state (pipeline, GST_STATE_NULL);

    g_print ("Deleting pipeline\n");
    gst_object_unref (GST_OBJECT (pipeline));
    g_main_loop_unref (loop);
    loop = NULL;
    return NULL;
}
static void stop_loop(){
    if (loop == NULL){
        return;
    }
    g_main_loop_quit(loop);
}

int main(int argc, char **argv) {
    int ret = 0;

    if (argc > 1){
        char daemon[] = "--daemon";
        int i = 0;
        while (daemon[i] && argv[1][i] && daemon[i] == argv[1][i]){
            i++;
        }
        if (i == 8 && !argv[1][i]){
            ret = init_daemon();
        }
    }

    if (!ret){
        int re, fd;
        struct timeval tv_start, tv_end;
        unlink( FIFO_FILE );
        umask(0);
        re = mkfifo( FIFO_FILE, 0777 );
		fd = open(FIFO_FILE, O_RDONLY|O_NONBLOCK);
        gettimeofday(&tv_start, NULL);
        tv_end = tv_start;
        while (1){
            char n, cmd[256];
            n = read(fd, cmd, sizeof(cmd));
            if (n > 0){
                printf("n=%d, cmd=%s\n", n, cmd);
                switch (cmd[0]){
                    case '0':
                        stop_loop();
                        break;
                    case 'r':
                        if (n > 8 && !memcmp(cmd, "rtmp://", 7)){
                            pthread_t id;
                            GstElement *videosrc, *text, *videoenc, *videoconvert, *muxer, *sink;
                            /* Initialisation */
                            {
                                if (loop != NULL){
                                    break;
                                }
                                gst_init (NULL, NULL);
                                loop = g_main_loop_new (NULL, FALSE);
                            }

                            /* Create gstreamer elements */
                            pipeline = gst_pipeline_new ("media-player");
                            videosrc = gst_element_factory_make ("imxv4l2videosrc",         "video-camrta-source");
                            text     = gst_element_factory_make ("textoverlay",         "text");
                            videoenc = gst_element_factory_make ("imxvpuenc_h264",  "video-h264-byte-stream");
                            videoconvert = gst_element_factory_make ("h264parse",       "video-convert");
                            muxer    = gst_element_factory_make ("flvmux",          "flv-muxer");
                            sink     = gst_element_factory_make ("rtmpsink",      "sink");

                            if (!pipeline || !videosrc || !text || !videoenc || !videoconvert || !muxer || !sink) {
                                g_printerr ("One element could not be created. Exiting.\n");
                                loop = NULL;
                                break;
                            }
                            //"rtmp://25582.lsspublish.aodianyun.com/server0/stream"
                            /* Set up the pipeline */
                            /* we set the input filename to the source element */
                            g_object_set (G_OBJECT (sink), "location", cmd, NULL);
                            //g_object_set (G_OBJECT (text), "text", cmd, NULL);
                            /* we add a message handler */

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
                            gst_element_set_state (pipeline, GST_STATE_PLAYING);
                            pthread_create(&id, NULL, (void*)start_loop, NULL);
                        }
                        break;
                    default:
                        break;
                }
                gettimeofday(&tv_start, NULL);
            }
            else{
                gettimeofday(&tv_end, NULL);
                if ((tv_end.tv_sec - tv_start.tv_sec) > 10 && loop != NULL){
                    printf("time out! stop loop\n");
                    stop_loop();
                }
                else{
                    printf("no msg %d, %d\n", n, (int)(tv_end.tv_sec - tv_start.tv_sec));
                }
                sleep(1);
            }
        }
    }
    return ret;
}
