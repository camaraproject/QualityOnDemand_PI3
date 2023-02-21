# Build stage
FROM golang:1.19.3-alpine AS build

WORKDIR /src
COPY . .

RUN apk add --no-cache make git 
RUN make

# Execution stage
FROM alpine:3.17

WORKDIR /root
COPY --from=build /src/qodservice /root
COPY --from=build /src/config /root/config

RUN apk add gcompat

CMD ["./qodservice"]

