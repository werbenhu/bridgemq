
# bridgemq
The fully compliant, embeddable high-performance Go MQTT broker server, support bridge mode.

Bridgemq can be used as a standalone broker and it based on [mochi-co/mqtt](https://github.com/mochi-co/mqtt) . bridgemq also support bridge mode, it use [hashicorp/serf](https://github.com/hashicorp/serf) for discovery and [grpc](https://github.com/grpc/grpc-go) to transmit message between agents .


## Quick Start
Simply checkout this repository and run the [cmd/main.go](cmd/main.go) entrypoint in the [cmd](cmd) folder which will expose tcp (:1883), tls tcp(:8883), websocket(:8085), and dashboard(:8080) listeners.


#### Usage
```
Usage of bridgemq:
  -agent-addr string
        listening addr for bridge agent, such as 192.168.0.1:7933 or :7933 (default ":7933")
  -agent-advertise string
        address to advertise to other agent. used for nat traversal. such as 192.168.0.1:7933 or www.xxx.com:7933
  -agent-name string
        the name of current agent, this parameter is not set, a name is randomly generated
  -agents string
        seeds list of bridge member agents, such as 192.168.0.1:7933,192.168.0.2:7933
  -bridge
        optional value for bridge mode
  -dashboard string
        http port for web info dashboard listener, if this parameter is not set, this default port is 8080 (default "8080")
  -pipe-port string
        transmit port (grpc server) to receive msg from other bridge agent. such as 8933 (default "8933")
  -tcp string
        network port for mqtt tcp listener
  -tls string
        network port for mqtt tls listener, if this parameter is not set, the service will not open, if set this then parameter -tls-ca, -tls-cert and -tls-key must be set
  -tls-ca string
        ca file path for tls listener
  -tls-cert string
        certificate file path for tls listener
  -tls-key string
        key file path for tls listener
  -ws string
        network port for mqtt websocket listener, if this parameter is not set, this service will not open
```

#### Simple start 
```sh
cd cmd
go build -o bridgemq && ./bridgemq
```

#### Start with websocket
```sh
./bridgemq -ws=8085
```

#### Start with dashboard
```sh
./bridgemq -dashboard=8080
```

#### Start with TLS
```sh
# only tls on 8883
./bridgemq -tls=8883 -tls-ca="./ca.crt" -tls-cert="./server.crt" -tls-key="./server.key"

# start both tcp and tls tcp
./bridgemq -tcp=1883 -tls=8883 -tls-ca="./ca.crt" -tls-cert="./server.crt" -tls-key="./server.key"
```

#### Start bridge mod
```sh
# start the first node
./bridgemq -tcp=1883 \
  -bridge \
  -agent-addr=":7933" \
  -agent-advertise="192.168.1.10:7933" \
  -pipe-port=8933

# start the second node
./bridgemq -tcp=1884 \
  -bridge \
  -agent-addr=":7934" \
  -agent-advertise="192.168.1.10:7934" \
  -agents="192.168.1.10:7933" \
  -pipe-port=8934

```

### Using Docker
A simple Dockerfile is provided for running the [cmd/main.go](cmd/main.go) Websocket, TCP, and Stats server:

```sh
docker build -t werbenhu/bridgemq:latest .
docker run \
    -e TCP_PORT=1883 \
    -v  /root/bridgemq/log:/log \
    -v  /root/bridgemq/data:/data \
    -p 1883:1883 \
    --name bridgemq \
    -d werbenhu/bridgemq
```

#### Docker start tls
```sh
docker run \
    -e TLS_PORT=8883 \
    -e TLS_CA=/ssl/root.crt \
    -e TLS_CERT=/ssl/server.crt \
    -e TLS_KEY=/ssl/server.key \
    -v /root/bridgemq/log:/log \
    -v /root/bridgemq/ssl:/ssl \
    -v /root/bridgemq/data:/data \
    -p 8883:8883 \
    --name bridgemq \
    -d werbenhu/bridgemq
```

#### Docker start cluster
```sh
# start node1
docker run \
    -e TCP_PORT=1883 \
    -e BRIDGE=true \
    -e AGENT_ADDR=":7933" \
    -e AGENT_ADVERTISE="172.16.3.3:7933" \
    -e PIPE_PORT="8933" \
    -v /root/bridgemq/log1:/log \
    -v /root/bridgemq/data1:/data \
    --name node1 \
    -p 1883:1883 \
    -p 8933:8933 \
    -p 7933:7933 \
    -p 7933:7933/udp \
    -d werbenhu/bridgemq

# start node2
docker run \
    -e TCP_PORT=1884 \
    -e BRIDGE=true \
    -e AGENT_ADDR=":7934" \
    -e AGENT_ADVERTISE="172.16.3.3:7934" \
    -e PIPE_PORT="8934" \
    -e AGENTS="172.16.3.3:7933" \
    -v /root/bridgemq/log2:/log \
    -v /root/bridgemq/data2:/data \
    --name node2 \
    -p 1884:1884 \
    -p 8934:8934 \
    -p 7934:7934 \
    -p 7934:7934/udp \
    -d werbenhu/bridgemq
```

## Contributions
Contributions and feedback are both welcomed and encouraged! Open an [issue](https://github.com/werbenhu/bridgemq/issues) to report a bug, ask a question, or make a feature request.

## Developing && Performance Benchmarks
Refer to [mochi-co/mqtt](https://github.com/mochi-co/mqtt)



