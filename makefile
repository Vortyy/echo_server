all : server

client :
	go compile 
server :
	clang server.c -o server
clean : 
	rm server
