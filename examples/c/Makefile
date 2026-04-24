CC      ?= gcc
CFLAGS  ?= -Wall -Wextra -O2 -std=c11
LDLIBS  ?= -lcurl

.PHONY: all clean run

all: demo

demo: demo.c mgt_client.c mgt_client.h
	$(CC) $(CFLAGS) -o demo mgt_client.c demo.c $(LDLIBS)

run: demo
	MGT_BASE=$${MGT_BASE:-http://localhost:3000} \
	MGT_USER=$${MGT_USER:-admin} \
	MGT_PASS=$${MGT_PASS:-admin} \
	./demo

clean:
	rm -f demo *.o
