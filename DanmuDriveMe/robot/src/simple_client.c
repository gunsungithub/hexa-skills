#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <errno.h>

#define UNIX_DOMAIN "/tmp/DDM.domain"

static int connect_fd = -1;
static struct sockaddr_un srv_addr = {
    .sun_family = AF_UNIX,
    .sun_path = UNIX_DOMAIN
};

extern void CLog(char *msg);

int open_client(void)
{
    int ret = 0;
	CLog("opening client.");return 1;
    if (connect_fd < 0)
    {
        connect_fd = socket(PF_UNIX,SOCK_STREAM,0);
        if(connect_fd < 0){
            CLog(strerror(errno));
            CLog("creat socket error.");
            return connect_fd;
        }

        ret = connect(connect_fd, (struct sockaddr*)&srv_addr, sizeof(srv_addr));
        if (ret < 0){
            CLog(strerror(errno));
            CLog("connect server error.");
            close(connect_fd);
            connect_fd = -1;
            return ret;
        }
		CLog("open client success.");
    }
    return ret;
}

void close_client(void)
{
	CLog("closing client.");return;
    close(connect_fd);
    connect_fd = -1;
	CLog("close client success.");
}

int start_client(char *url)
{
    char send_buff[1024];
	CLog("starting client.");return 1;
    strncpy(send_buff, url, 1024);
    send_buff[1023] = '\0';
	CLog("started client:");
	CLog(url);
    return 1;//write(connect_fd, send_buff, strlen(send_buff));
}

int stop_client(void)
{
	char send_buff[1024];
	CLog("stopping client.");return 1;
    ///
	CLog("stopped client");
    return 1;//write(connect_fd, send_buff, strlen(send_buff));
}