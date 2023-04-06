FROM golang:1.19.0-alpine3.15 AS builder

WORKDIR /app

COPY . ./

RUN go build -o /app/bridgemq ./cmd


FROM alpine

ENV TCP_PORT=""
ENV TLS_PORT=""
ENV TLS_CA=""
ENV TLS_CERT=""
ENV TLS_KEY=""
ENV WEBSOCKS=""
ENV DASHBOARD=""
ENV BRIDGE=false
ENV AGENTS=""
ENV AGENT_NAME=""
ENV AGENT_ADDR=":7933"
ENV AGENT_ADVERTISE=""
ENV PIPE_PORT=""


WORKDIR /
COPY --from=builder /app/bridgemq .

ENTRYPOINT [ "/bin/sh", "-c", "/bridgemq \ 
    -tcp=${TCP_PORT} \ 
    -tls=${TLS_PORT} \ 
    -tls-ca=${TLS_CA} \ 
    -tls-cert=${TLS_CERT} \ 
    -tls-key=${TLS_KEY} \ 
    -dashboard=${DASHBOARD} \ 
    -ws=${WEBSOCKS} \ 
    -bridge=${BRIDGE} \ 
    -agent-name=${AGENT_NAME} \ 
    -agent-addr=${AGENT_ADDR} \ 
    -agent-advertise=${AGENT_ADVERTISE} \ 
    -agents=${AGENTS} \
    -pipe-port=${PIPE_PORT}" \ 
    ]
