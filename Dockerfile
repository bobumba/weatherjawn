FROM alpine:latest AS sqlcreate

RUN apk add --no-cache sqlite

RUN sqlite3 /tmp/weatherjawn.db "CREATE TABLE airfeelings (id INTEGER PRIMARY KEY AUTOINCREMENT, datetime DATETIME, temperature REAL, humidity REAL, barometricpressure REAL);"

# Use the official Golang image as the base image
FROM golang:1.22

# Set the Current Working Directory inside the container
WORKDIR /app


# Copy go mod and sum filesk
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build weather.go

COPY --from=sqlcreate /tmp/weatherjawn.db /app/weatherjawn.db

# Expose port 8080 to the outside world
EXPOSE 51101

# Command to run the executable
CMD ["/app/weather"]

