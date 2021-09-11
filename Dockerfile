FROM ubuntu:20.04

WORKDIR /usr/app

COPY ./botmsg .
COPY ./bot-config.json .
# COPY ./components/ /usr/app/
CMD ["/usr/app/botmsg"]

# EXPOSE 4000