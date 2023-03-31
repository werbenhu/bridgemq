package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/hooks/auth"
	"github.com/mochi-co/mqtt/v2/hooks/storage/bolt"
	"github.com/mochi-co/mqtt/v2/listeners"
	"github.com/rs/zerolog"
	"github.com/werbenhu/bridgemq"
	"go.etcd.io/bbolt"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newTlsConfig(pemFile string, keyFile string, caFile string) *tls.Config {
	pem, err := tls.LoadX509KeyPair(pemFile, keyFile)
	if err != nil {
		log.Fatalf("[ERROR] tls serve LoadX509KeyPair cert file:%s Err:%s\n", pemFile, err.Error())
	}

	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("[ERROR] tls serve read ca file failed. err:%s", err.Error())
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca)
	if !caPool.AppendCertsFromPEM(ca) {
		log.Fatalf("[ERROR] tls serve append cert pool failed.\n")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{pem},
		ClientCAs:          caPool,
		InsecureSkipVerify: false,
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}
}

func main() {
	tcpPort := flag.String("tcp", "", "network port for mqtt tcp listener")
	tlsPort := flag.String("tls", "", "network port for mqtt tls listener, if this parameter is not set, the service will not open, if set this then parameter -tls-ca, -tls-cert and -tls-key must be set")
	wsPort := flag.String("ws", "", "network port for mqtt websocket listener, if this parameter is not set, this service will not open")
	tlsCa := flag.String("tls-ca", "", "ca file path for tls listener")
	tlsCert := flag.String("tls-cert", "", "certificate file path for tls listener")
	tlsKey := flag.String("tls-key", "", "key file path for tls listener")
	dashboard := flag.String("dashboard", "8080", "http port for web info dashboard listener, if this parameter is not set, this default port is 8080")

	isBridge := flag.Bool("bridge", false, "optional value for bridge mode")
	agents := flag.String("agents", "", "seeds list of bridge member agents, such as 192.168.0.1:7933,192.168.0.2:7933")
	agentName := flag.String("agent-name", "", "the name of current agent, this parameter is not set, a name is randomly generated")
	agentAddr := flag.String("agent-addr", ":7933", "listening addr for bridge agent, such as 192.168.0.1:7933 or :7933")
	agentAdvertise := flag.String("agent-advertise", "", "address to advertise to other agent. used for nat traversal. such as 192.168.0.1:7933 or www.xxx.com:7933")
	pipePort := flag.String("pipe-port", "8933", "transmit port (grpc server) to receive msg from other bridge agent. such as 8933")

	flag.Parse()
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	writers := io.MultiWriter(&lumberjack.Logger{
		Filename:   "./log/bridgemq.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
	}, os.Stderr)

	logger := zerolog.New(writers).With().Timestamp().Logger().Level(zerolog.InfoLevel).Output(zerolog.ConsoleWriter{
		Out:        writers,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    true,
	})

	server := mqtt.New(&mqtt.Options{
		Logger: &logger,
	})

	os.MkdirAll("./data", fs.ModePerm)
	_ = server.AddHook(new(auth.AllowHook), nil)
	_ = server.AddHook(new(bolt.Hook), &bolt.Options{
		Path: "./data/bolt.db",
		Options: &bbolt.Options{
			Timeout: 500 * time.Millisecond,
		},
	})

	// if both tcp port and tsl port not set, default open tcp service on 1883
	if *tcpPort == "" && *tlsPort == "" {
		*tcpPort = "1883"
	}

	if *tcpPort != "" {
		tcp := listeners.NewTCP("t1", ":"+*tcpPort, nil)
		err := server.AddListener(tcp)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *tlsPort != "" {
		tlsConfig := newTlsConfig(*tlsCert, *tlsKey, *tlsCa)
		tlsTcp := listeners.NewTCP("tls1", ":"+*tlsPort, &listeners.Config{
			TLSConfig: tlsConfig,
		})
		err := server.AddListener(tlsTcp)
		if err != nil {
			log.Fatal(err)
		}
	}

	// if bridge mode on, add bridge hook to mqtt server
	if *isBridge {
		_ = server.AddHook(new(bridgemq.Hook), []bridgemq.IOption{
			bridgemq.OptName(*agentName),
			bridgemq.OptAddr(*agentAddr),
			bridgemq.OptAgents(*agents),
			bridgemq.OptBroker(server),
			bridgemq.OptAdvertise(*agentAdvertise),
			bridgemq.OptPipePort(*pipePort),
		})
	}

	// if websocket port not set, do not open the ws service
	if *wsPort != "" {
		ws := listeners.NewWebsocket("ws1", ":"+*wsPort, nil)
		err := server.AddListener(ws)
		if err != nil {
			log.Fatal(err)
		}
	}

	// if http port not set, do not open the http service
	if *dashboard != "" {
		stats := listeners.NewHTTPStats("stats", ":"+*dashboard, nil, server.Info)
		err := server.AddListener(stats)
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn().Msg("caught signal, stopping...")
	server.Close()
	server.Log.Info().Msg("main.go finished")
}
