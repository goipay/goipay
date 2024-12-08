package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/chekist32/goipay/internal/dto"
	handler_v1 "github.com/chekist32/goipay/internal/handler/v1"
	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/chekist32/goipay/internal/processor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

type CliOpts struct {
	ConfigPath        string
	ClientCAPaths     string
	ReflectionEnabled bool
}

type TlsMode string

const (
	NONE_TLS_MODE TlsMode = "none"
	TLS_TLS_MODE  TlsMode = "tls"
	MTLS_TLS_MODE TlsMode = "mtls"
)

type AppConfigDaemon struct {
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type AppConfigTls struct {
	Mode string `yaml:"mode"`
	Ca   string `yaml:"ca"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type AppConfig struct {
	Server struct {
		Host string       `yaml:"host"`
		Port string       `yaml:"port"`
		Tls  AppConfigTls `yaml:"tls"`
	} `yaml:"server"`

	Database struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		Name string `yaml:"name"`
	} `yaml:"database"`

	Coin struct {
		Xmr struct {
			Daemon AppConfigDaemon `yaml:"daemon"`
		} `yaml:"xmr"`
	} `yaml:"coin"`
}

func NewAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf AppConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	conf.Server.Host = os.ExpandEnv(conf.Server.Host)
	conf.Server.Port = os.ExpandEnv(conf.Server.Port)

	conf.Server.Tls.Mode = os.ExpandEnv(conf.Server.Tls.Mode)
	conf.Server.Tls.Ca = os.ExpandEnv(conf.Server.Tls.Ca)
	conf.Server.Tls.Cert = os.ExpandEnv(conf.Server.Tls.Cert)
	conf.Server.Tls.Key = os.ExpandEnv(conf.Server.Tls.Key)

	conf.Database.Host = os.ExpandEnv(conf.Database.Host)
	conf.Database.Port = os.ExpandEnv(conf.Database.Port)
	conf.Database.User = os.ExpandEnv(conf.Database.User)
	conf.Database.Pass = os.ExpandEnv(conf.Database.Pass)
	conf.Database.Name = os.ExpandEnv(conf.Database.Name)

	conf.Coin.Xmr.Daemon.Url = os.ExpandEnv(conf.Coin.Xmr.Daemon.Url)
	conf.Coin.Xmr.Daemon.User = os.ExpandEnv(conf.Coin.Xmr.Daemon.User)
	conf.Coin.Xmr.Daemon.Pass = os.ExpandEnv(conf.Coin.Xmr.Daemon.Pass)

	return &conf, nil
}

type App struct {
	ctxCancel context.CancelFunc

	config *AppConfig
	opts   *CliOpts
	log    *zerolog.Logger

	dbConnPool       *pgxpool.Pool
	paymentProcessor *processor.PaymentProcessor
}

func (a *App) Start(ctx context.Context) error {
	if err := a.dbConnPool.Ping(ctx); err != nil {
		a.log.Err(err).Msg("failed to connect to database")
		return err
	}
	defer a.dbConnPool.Close()

	lis, err := net.Listen("tcp", a.config.Server.Host+":"+a.config.Server.Port)
	if err != nil {
		a.log.Fatal().Msgf("failed to listen on port %v: %v", a.config.Server.Port, err)
	}

	g := grpc.NewServer(getGrpcServerOptions(a)...)
	pb_v1.RegisterUserServiceServer(g, handler_v1.NewUserGrpc(a.dbConnPool, a.log))
	pb_v1.RegisterInvoiceServiceServer(g, handler_v1.NewInvoiceGrpc(a.dbConnPool, a.paymentProcessor, a.log))

	if a.opts.ReflectionEnabled {
		reflection.Register(g)
	}

	ch := make(chan error, 1)
	go func() {
		if err := g.Serve(lis); err != nil {
			a.log.Err(err).Msg("failed to start server")
			ch <- err
		}
		close(ch)
	}()

	a.log.Info().Msgf("Starting server %v\n", lis.Addr())

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		a.ctxCancel()
		g.GracefulStop()
		return nil
	}
}

func getGrpcServerOptions(a *App) []grpc.ServerOption {
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			NewMetadataInterceptor(a.log).Intercepte,
			NewRequestLoggingInterceptor(a.log).Intercepte,
		),
	}

	if creds, enabled := getGrpcCrednetials(a.config, a.opts); enabled {
		grpcOpts = append(grpcOpts, creds)
	}

	return grpcOpts
}

func getMtlsCofig(c *AppConfig, opts *CliOpts) *tls.Config {
	config := getTlsConfig(c)

	if strings.TrimSpace(opts.ClientCAPaths) == "" {
		log.Fatal("-client-ca must specify at least one path")
	}

	paths := strings.Split(strings.TrimSpace(opts.ClientCAPaths), ",")
	if len(paths) == 0 {
		log.Fatal("-client-ca must specify at least one path")
	}

	certPool := x509.NewCertPool()
	for i := 0; i < len(paths); i++ {
		trustedCert, err := os.ReadFile(paths[i])
		if err != nil {
			log.Fatalf("Failed to load trusted client certificate %v", err)
		}
		if !certPool.AppendCertsFromPEM(trustedCert) {
			log.Fatalf("Failed to append trusted client certificate %v to certificate pool", paths[i])
		}
	}

	config.ClientCAs = certPool
	config.ClientAuth = tls.RequireAndVerifyClientCert

	return config
}

func getTlsConfig(c *AppConfig) *tls.Config {
	serverCert, err := tls.LoadX509KeyPair(c.Server.Tls.Cert, c.Server.Tls.Key)
	if err != nil {
		log.Fatalf("Failed to load server certificate and key %v", err)
	}

	trustedCert, err := os.ReadFile(c.Server.Tls.Ca)
	if err != nil {
		log.Fatalf("Failed to load trusted server certificate %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		log.Fatalf("Failed to append trusted server certificate %v to certificate pool", c.Server.Tls.Ca)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      certPool,
	}

	return tlsConfig
}

func getGrpcCrednetials(c *AppConfig, opts *CliOpts) (grpc.ServerOption, bool) {
	mode := TlsMode(c.Server.Tls.Mode)

	switch mode {
	case NONE_TLS_MODE:
		return nil, false
	case TLS_TLS_MODE:
		return grpc.Creds(credentials.NewTLS(getTlsConfig(c))), true
	case MTLS_TLS_MODE:
		return grpc.Creds(credentials.NewTLS(getMtlsCofig(c, opts))), true
	default:
		log.Fatalf("Invalid TLS mode: %v. It must be one of: none, tls, mtls", mode)
	}

	return nil, false
}

func appConfigToDaemonsConfig(c *AppConfig) *dto.DaemonsConfig {
	acdTodc := func(c *AppConfigDaemon) *dto.DaemonConfig {
		return &dto.DaemonConfig{
			Url:  c.Url,
			User: c.User,
			Pass: c.Pass,
		}
	}

	return &dto.DaemonsConfig{
		Xmr: *acdTodc(&c.Coin.Xmr.Daemon),
	}
}

func getLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger()
	return &logger
}

func NewApp(opts CliOpts) *App {
	ctx, cancel := context.WithCancel(context.Background())
	log := getLogger()

	conf, err := NewAppConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal().Err(err)
	}

	dbUrl := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", conf.Database.User, conf.Database.Pass, conf.Database.Host, conf.Database.Port, conf.Database.Name)
	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal().Err(err)
	}

	pp, err := processor.NewPaymentProcessor(ctx, connPool, appConfigToDaemonsConfig(conf), log)
	if err != nil {
		log.Fatal().Err(err)
	}

	return &App{
		log:              log,
		ctxCancel:        cancel,
		opts:             &opts,
		config:           conf,
		dbConnPool:       connPool,
		paymentProcessor: pp,
	}
}
