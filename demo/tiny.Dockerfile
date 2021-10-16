FROM ubuntu as build-env

RUN apt-get update
RUN apt-get install -y wget tar
RUN wget https://github.com/osa1/tiny/releases/download/v0.9.0/tiny-ubuntu-20.04.tar.gz
RUN tar -xvzf tiny-ubuntu-20.04.tar.gz

FROM ubuntu
COPY --from=build-env /tiny /tiny
CMD ["sleep", "infinity"]
