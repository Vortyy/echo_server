/* Implement Epoll event sequence
 * Epoll is api provide by kernels to write a non-blocking IO application (also called async now)
 * see man epoll or io_during (called kqueue in BSD, IOCP in Windows) */

//Simple socket server system --> done
//Epoll event manager

#include <string.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <asm-generic/socket.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>

#define PORT 8080     /* listening port */
#define IP INADDR_ANY /* all the listening addr */
#define BACKLOG 3     /* max pending connection to the socket */

/* error_crit: write an error msg and exit the process */
void error_crit(char *msg){
  fprintf(stderr, "error: %s\n", msg);
  exit(1);
}

int main(int argc, char **argv){
  struct sockaddr_in addr = {.sin_family=AF_INET, .sin_addr=IP, .sin_port=htons(PORT)}; /* set my address */

  int fdsocket;
  int opt = 1;
  char buffer[BUFSIZ];

  if((fdsocket = socket(AF_INET, SOCK_STREAM, 0)) == -1) /* check if fd link to socket is init */
    error_crit("can't initialise fd for the socket");

  printf("socket fd : %d\n", fdsocket);

  if(setsockopt(fdsocket, SOL_SOCKET, SO_REUSEADDR | SO_REUSEPORT, &opt, sizeof(opt)) != 0)
    error_crit("socket option can't be set...");
  if(bind(fdsocket, (struct sockaddr *) &addr, sizeof(struct sockaddr_in)) != 0)
    error_crit("socket can't be bind to the address");

  printf("socket bind INADDR_ANY:%d\n", PORT); /* socket binded */ 

  if(listen(fdsocket, BACKLOG) != 0)
    error_crit("socket can't listen...");

  printf("socket listening...\n");

  int fdclient;
  struct sockaddr_in client_addr;

  socklen_t client_addr_size = sizeof(struct sockaddr); 

  while(1){
    fdclient = accept(fdsocket, (struct sockaddr *) &client_addr, &client_addr_size); /* blocking current thread is stoped */
    printf("Client connected with a fd: %d\n", fdclient);
    while(fdclient != -1){ /* a client is connected */
      int nread = recv(fdclient, buffer, BUFSIZ, 0);
      if(nread <= 0){ /* connection closed or error */
        printf("nread code : %d -> Closing client : %d\n", nread, fdclient);
        close(fdclient);
        fdclient = -1;
        break;
      }
      //char ip[INET_ADDRSTRLEN]; /* 15 + '\0' -> goes to 16 */
      //inet_ntop(AF_INET, &(client_addr.sin_addr), ip, INET_ADDRSTRLEN);
      printf("bytes receives : %2d, client connected on port : --> %4d\n", nread, client_addr.sin_port);
      send(fdclient, buffer, nread, 0);
    }
  }

  printf("closing fd socket...\n");
  close(fdsocket);
  return 0;
}
