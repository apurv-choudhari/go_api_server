FROM golang:1.23 AS builder

RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

WORKDIR /app/go-service
RUN go mod download
RUN go build -o go-app .

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    mariadb-server \
    supervisor \
    vim nano curl wget git unzip sudo net-tools telnet && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/go-service /app/go-service

COPY . .

RUN mkdir -p /var/run/mysqld && \
    chown root:mysql /var/run/mysqld && \
    chmod 774 /var/run/mysqld && \
    mkdir -p /docker-entrypoint-initdb.d/

COPY init.sql /docker-entrypoint-initdb.d/
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

CMD ["/usr/bin/supervisord"]