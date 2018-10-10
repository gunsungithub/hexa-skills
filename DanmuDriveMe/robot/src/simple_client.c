#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <errno.h>
#include <stddef.h>
#include <string.h>

static char msg[256];
#define printf(fmt, args...) do {snprintf(msg, 256, fmt, ##args);CLog(msg);}while(0)
extern void CLog(char *msg);

void setup(char *v){
    int n;
    int fd;
    printf("opening /tmp/myfifo\n");
    fd = open("/tmp/myfifo", O_WRONLY|O_NONBLOCK);
    printf("opened /tmp/myfifo fd=%d\n", fd);
    if (fd >= 0){
        printf("cmd:%s\n", v);
        n = write(fd, v, (strlen(v) + 1) * sizeof(char));
        printf("%d writed\n", n);
        if (n < 0){
            printf("[%s] write error.\n", strerror(errno));
        }
        close(fd);
    }
    else {
        printf("[%s] open for write error.\n", strerror(errno));
    }
}