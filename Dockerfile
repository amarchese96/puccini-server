FROM golang:1.18-buster AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /puccini-server

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build /puccini-server /puccini-server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/puccini-server"]