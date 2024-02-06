# Start from the official golang image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Install Templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Tailwind CLI Tool
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN mv tailwindcss-linux-x64 tailwindcss


# Build the Go app
RUN ./tailwindcss -i public/css/input.css -o public/css/output.css --minify
RUN templ generate
RUN go build -buildvcs=false -o main .

# Expose port 8080 to the outside world
EXPOSE 443

# Command to run the executable
CMD ["./main"]
