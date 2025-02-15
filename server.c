/***************************************************************************
 * Echo Server
 *
 * Just simple echo server which handle multiple connection to learn 
 * Poll api provide by kernels to write a non-blocking IO application 
 * In this example the server handles 4 clients simultanously (could handle 
 * lot more but i choose 4 'testing purpose')
 *
 * REF :
 * - man poll or io_during (called kqueue in BSD, IOCP in Windows) 
 **************************************************************************/

#include <signal.h>
#include <string.h>
#include <poll.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <asm-generic/socket.h>
#include <sys/poll.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>

#define NFDS (sizeof fds / sizeof(fds[0]))

#define SERVER 0      /* fds server idx */
#define PORT 8080     /* listening port */
#define IP INADDR_ANY /* all the listening addr */
#define BACKLOG 3     /* max pending connection to the socket */
#define MAX_FD 2      /* max fd open and checked for poll */

struct pollfd fds[MAX_FD]; /* list of fd checked */
static volatile sig_atomic_t keep_running = 1; /* boolean used in while loop for the server */

/* interruption_handler: is an handler that catch a ctrl+c (a.k.a SIGINT) and stop the while */
void interruption_handler(int signo){
  keep_running = 0;
}

/* error_crit: write an error msg and exit the process */
void error_crit(char *msg){
  fprintf(stderr, "error: %s\n", msg);
  exit(1);
}

/* get_index: get first index not used in fds or -1 if its full */
int get_idx(){
  int j = 1; /* because 0 is the server idx in fds */
  while(fds[j].fd != 0 && j < MAX_FD)
    j++;
  return (j == MAX_FD) ? -1 : j;
}

/* reset_pollfd: with an list fds and index inside this list reset the specific index */
void reset_pollfd(int idx){
  if(idx < 1 || idx > NFDS - 1)
    error_crit("idx is out of bound of fds");

  fds[idx].fd = 0;
  fds[idx].events = 0;
  fds[idx].revents = 0;
}

int main(int argc, char **argv){

  /* signal action creation */
  struct sigaction act = { 0 };
  act.sa_handler = interruption_handler;

  /* here (signo, act, oldact) -> oldact is usefull to keep the previous interruption 
   * action to be able to switch from the new one to old one in the program 
   * (cf. https://stackoverflow.com/questions/3635221/what-is-the-use-of-second-structureoldact-in-sigaction */
  sigaction(SIGINT, &act, NULL);

  struct sockaddr_in addr = {.sin_family=AF_INET, .sin_addr=IP, .sin_port=htons(PORT)}; /* set my address */

  int opt;

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

  printf("fds starting status: \n");
  for(int i=0; i < MAX_FD; i++){
    printf(" • fd=%d && events = 0x%x \n", fds[i].fd, fds[i].events);
  }

  char buffer[BUFSIZ]; /* that serve to read input from client */
  /* listening loop */
  while(keep_running){
    ready = poll(fds, MAX_FD, -1);
    
    if(ready > 0)
      printf(" Ready: %d \n", ready);

    for(int i = 0; i < MAX_FD; i++){
      if(fds[i].revents != 0){
        printf(" fd=%d events: %x \n", fds[i].fd, fds[i].revents); 
        if(i == SERVER){        /* server case */
          int fds_i = get_idx();
          int fdclient = accept(fds[SERVER].fd, (struct sockaddr *) &client_addr, (socklen_t *) &client_addr_size);

          if(fds_i == -1){      /* client full */
            write(fdclient, "server is full...\n", strlen("server is full...\n"));
            close(fdclient);
          } else {              /* adding new client */
            printf("client accepted on fd = %d\n", fdclient);
            
            fds[fds_i].fd = fdclient;
            fds[fds_i].events = POLLIN;
          }
        }
        else {                 /* client case */
          int n = read(fds[i].fd, buffer, BUFSIZ);
          if(n > 0){
            printf(" %d bytes read from %d \n", n, fds[i].fd); 
            int w = write(fds[i].fd, buffer, n);
          } else {
            printf(" Closing --> client fd=%d\n", fds[i].fd);
            close(fds[i].fd);
            reset_pollfd(i);
          }
        }
      }
    }
  }

  /* closing opened clients */
  printf("\nClosing opened clients...\n"); // \n because SIGINT not \n
  for (int i = 1; i < MAX_FD; i++)
    if(fds[i].fd != 0)
      close(fds[i].fd);

  /* closing server */
  printf("Closing server...\n");
  close(fds[SERVER].fd);
  return 0;
}
