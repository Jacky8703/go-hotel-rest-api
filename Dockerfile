FROM golang:1.25
WORKDIR /hotel
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-golang-hotel
EXPOSE 8080
CMD ["/docker-golang-hotel"]