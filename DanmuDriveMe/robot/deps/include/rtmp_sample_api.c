#include <rtmp.h>
#include <log.h>
static RTMP *rtmp=NULL;

int rtmp_sample_init(){
    rtmp=RTMP_Alloc();
    RTMP_Init(rtmp);
    rtmp->Link.timeout=5;
    return 0;
}

int rtmp_sample_connect(char *url){
    if(!RTMP_SetupURL(rtmp, url))
    {
        RTMP_Log(RTMP_LOGERROR, "SetupURL Err\n");
        RTMP_Free(rtmp);
        return -1;
    }
    RTMP_EnableWrite(rtmp);
    if (!RTMP_Connect(rtmp, NULL)){
        RTMP_Log(RTMP_LOGERROR, "Connect Err\n");
        RTMP_Free(rtmp);
        return -1;
    }
    if (!RTMP_ConnectStream(rtmp, 0)){
        RTMP_Log(RTMP_LOGERROR, "ConnectStream Err\n");
        RTMP_Close(rtmp);
        RTMP_Free(rtmp);
        return -1;
    }
    return 0;
}
int rtmp_sample_add_data(char *pbuf, int datalength){
    if (!RTMP_IsConnected(rtmp)){
        RTMP_Log(RTMP_LOGERROR,"rtmp is not connect\n");
        return -1;
    }
    if (!RTMP_Write(rtmp, pbuf, datalength)){
        RTMP_Log(RTMP_LOGERROR, "Rtmp Write Error\n");
        return -1;
    }
    return 0;
}
int rtmp_sample_disconnect(){
    if (rtmp != NULL){
        RTMP_Close(rtmp);
    }
    return 0;
}
int rtmp_sample_final(){
    if (rtmp != NULL){
        RTMP_Free(rtmp);
        rtmp = NULL;
    }
    return 0;
}