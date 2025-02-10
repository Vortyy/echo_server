all :
	clang server.c -o server && ./server
clean : 
	rm server
