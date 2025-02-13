# build client and server and start running server
all : build_all
	./server

build_all : build_server build_client

# build client into client exec
build_client :
	go build client.go 

# build server 
build_server :
	clang server.c -o server

clean : 
	rm server client
