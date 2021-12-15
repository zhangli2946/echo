FROM zhangli2946/builder:latest AS builder
WORKDIR /src
ADD . .
RUN make server

FROM alpine:3.12
WORKDIR /var/app
COPY --from=builder /src/server .
CMD ["/var/app/server"]