package config

type GlobalConfig struct {
	ServerPort         int    `env:"SERVER_PORT,required"`
	ArchiverServerPort int    `env:"ARCHIVER_PORT,required"`
	PostgresConnection string `env:"POSTGRES_CONNECTION,required"`
	RabbitmqConnection string `env:"RABBITMQ_CONNECTION,required"`
	GrpcConnection     string `env:"GRPC_CONNECTION,required"`
}
