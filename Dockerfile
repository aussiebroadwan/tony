FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Download the dependencies
RUN go mod download

# Build the application 
RUN CGO_ENABLED=1 go build -o tony .

FROM alpine

# Metadata
LABEL description="The Aussie BroadWAN Discord Bot"
LABEL vendor="Aussie BroadWAN"
LABEL version="0.1.0"

# Copy the compiled binary from the builder stage
COPY --from=builder /app/tony /tony

# Run the bot
ENTRYPOINT ["/tony"]