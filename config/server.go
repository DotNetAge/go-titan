package config

type CertConfig struct {
	PublicKey  string `mapstructure:"pub"`  // PublicKey 返回公钥文件地址
	PrivateKey string `mapstructure:"key"` // PrivateKey 返回私钥文件地址
}

type ServerConfig struct {
	Name          string    `mapstructure:"name"`
	EndPoint      *EndPoint `mapstructure:"endpoint"`
	RegistryAddrs []string  `mapstructure:"registry"`
}

type GatewayConfig struct {
	Name          string      `mapstructure:"name"`
	EndPoint      *EndPoint   `mapstructure:"endpoint"`
	Transports    []*EndPoint `mapstructure:"trans"`
	RegistryAddrs []string    `mapstructure:"registry"`
	CORS          *CORSConfig `mapstructure:"cors"`
}
