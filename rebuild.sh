#!/bin/bash
go build -o botmsg *.go;
# docker stop botmsg && docker rm botmsg
docker rmi rogerbridge/botmsg:test;
docker build -t rogerbridge/botmsg:test .;
docker push rogerbridge/botmsg:test;