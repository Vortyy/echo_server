/* Implement Epoll event sequence
 * Epoll is api provide by kernels to write a non-blocking IO application (also called async now)
 * see man epoll or io_during (called kqueue in BSD, IOCP in Windows) */

//Simple socket server system --> done
//Epoll event manager

#include <poll.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <asm-generic/socket.h>
#include <sys/poll.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>

#define SERVER 0
#define PORT 8080     /* listening port */
#define IP INADDR_ANY /* all the listening addr */
#define BACKLOG 3     /* max pending connection to the socket */
#define MAX_FD 5      /* max fd open and checked for poll */

struct pollfd fds[MAX_FD]; /* list of fd checked */

/* error_crit: write an error msg and exit the process */
void error_crit(char *msg){
  fprintf(stderr, "error: %s\n", msg);
  exit(1);
}

int main(int argc, char **argv){
  struct sockaddr_in addr = {.sin_family=AF_INET, .sin_addr=IP, .sin_port=htons(PORT)}; /* set my address */

  int opt;
  int current_con = 1; /* always the server */
  char buffer[10];

  if((fds[SERVER].fd = socket(AF_INET, SOCK_STREAM | SOCK_NONBLOCK, 0)) == -1) /* check if fd link to socket is init and set it as non-blocking */
    error_crit("can't initialise fd for the socket");

  printf("socket fd : %d\n", fds[SERVER].fd);

  if(setsockopt(fds[SERVER].fd, SOL_SOCKET, SO_REUSEADDR | SO_REUSEPORT, &opt, sizeof(opt)) != 0)
    error_crit("socket option can't be set...");
  if(bind(fds[SERVER].fd, (struct sockaddr *) &addr, sizeof(struct sockaddr_in)) != 0)
    error_crit("socket can't be bind to the address");

  printf("socket bind INADDR_ANY:%d\n", PORT); /* socket binded */ 

  if(listen(fds[SERVER].fd, BACKLOG) != 0)
    error_crit("socket can't listen...");

  printf("socket listening...\n");

  int fdclient;
  struct sockaddr_in client_addr;
  socklen_t client_addr_size = sizeof(struct sockaddr); 

  fds[SERVER].events = POLLIN;
  int ready;

  while(1){
    ready = poll(fds, MAX_FD, -1);
    printf(" Ready: %d \n", ready);
    for(int i = 0; i < MAX_FD; i++){
      if(fds[i].revents != 0){
        if(i == SERVER){ /* server case */
          int fdclient = accept(fds[SERVER].fd, (struct sockaddr *) &client_addr, (socklen_t *) &client_addr_size);
          printf("client accepted on fd = %d\n", fdclient);
          fds[current_con].fd = fdclient;
          fds[current_con].events = POLLIN;
          current_con++;
        }
        else { /* client case */
          int n = read(fds[i].fd, buffer, 10);
          printf(" %d bytes read from %d \n", n, fds[i].fd); 
          int w = write(fds[i].fd, buffer, n);
        }
      }
    }
  }

  printf("closing fd socket...\n");
  close(fds[SERVER].fd);
  return 0;
}
