FROM alpine
COPY jwt /usr/local/bin/.
ENTRYPOINT ["jwt"]
