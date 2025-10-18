package defaults

type HttpSystemConfig struct {
	Enabled bool   `env:"ENABLED" envDefault:"true"`
	Port    uint16 `env:"PORT" envDefault:"8081"`
}

type HttpConfig struct {
	Enabled bool   `env:"ENABLED" envDefault:"true"`
	Host    string `env:"HOST" envDefault:""`
	Port    uint16 `env:"PORT" envDefault:"8080"`
}

type GrpcConfig struct {
	Enabled bool   `env:"ENABLED" envDefault:"true"`
	Host    string `env:"HOST" envDefault:""`
	Port    uint16 `env:"PORT" envDefault:"8000"`
}
