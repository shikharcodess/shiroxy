package models

type Config struct {
	Default  Default    `json:"default"`
	Frontend []Frontend `json:"frontend"`
	Backend  []Backend  `json:"backend"`
	Logging  Logging    `json:"logging"`
	Webhook  []Webhook  `json:"webhook"`
	Health   Health     `json:"health"`
}

type Storage struct {
	Location              string `json:"location"`
	RedisHost             string `json:"redis_host"`
	RedisPort             string `json:"redis_port"`
	RedisPassword         string `json:"redis_password"`
	RedisConnectionString string `json:"redis_connection_string"`
}

type Analytics struct {
	CollectionInterval int    `json:"collection_interval"`
	RouteName          string `json:"route_name"`
}

type Default struct {
	TIMEOUT struct {
		Connect string `json:"connect"`
		Server  string `json:"server"`
		Client  string `json:"client"`
	} `json:"timeout"`
	Mode                        string    `json:"mode"`
	Enable_Dns_Challenge_Solver bool      `json:"enable_dns_challenge_solver"`
	Storage                     Storage   `json:"storage"`
	Analytics                   Analytics `json:"analytics"`
}

type Frontend struct {
	Bind struct {
		Port   string `json:"port"`
		Host   string `json:"host"`
		Secure struct {
			Enable     bool   `json:"enable"`
			Target     string `json:"target"`
			CertAndKey struct {
				Cert string `json:"cert"`
				Key  string `json:"key"`
			} `json:"certandkey"`
			MultipleCertAndKeyLocation string `json:"multiple_certandkey_location"`
		} `json:"secure"`
	} `json:"bind"`
	Options          []string `json:"options"`
	Default_Backend  string   `json:"default_backend"`
	Fallback_Backend string   `json:"fallback_backend"`
}

type Backend struct {
	Name    string `json:"name"`
	Servers []struct {
		Id   string `json:"id"`
		Host string `json:"host"`
		Port string `json:"port"`
	}
	Health       bool   `json:"health"`
	Secure       bool   `json:"secure"`
	SecureVarify bool   `json:"secure_varify"`
	Balance      string `json:"balance"`
}

type Logging struct {
	Enable      bool `json:"enable"`
	EnableRmote bool `json:"enable_remote"`
	RemoteBind  struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"remote_bind"`
	Mode    string   `json:"mode"`
	Schema  []string `json:"schema"`
	Include []string `json:"include"`
}

type Webhook struct {
	Enable bool     `json:"enable"`
	Events []string `json:"events"`
	Url    string   `json:"url"`
}

type Health struct {
	Enable bool `json:"enable"`
	Bind   struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"bind"`
	Auth struct {
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"auth"`
}
