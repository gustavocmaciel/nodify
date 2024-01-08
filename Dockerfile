# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory in the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

RUN go mod tidy

# Install necessary packages in the container
RUN apt-get update && apt-get install -y \
    openjdk-17-jdk \
    graphviz \
    gnuplot \
    bzip2

# Download and unpack the tarball
RUN curl -L https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2 -o maelstrom.tar.bz2 \
    && tar -xjf maelstrom.tar.bz2 \
    && rm maelstrom.tar.bz2

CMD ["tail", "-f", "/dev/null"]
