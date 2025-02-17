FROM debian


RUN mkdir /build

COPY ./main /build/lvm-proxy

ENTRYPOINT ["/build/lvm-proxy"]

