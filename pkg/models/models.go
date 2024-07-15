package models

type Config struct {
	Default  Default   `json:"default"`
	Frontend Frontend  `json:"frontend"`
	Backend  Backend   `json:"backend"`
	Logging  Logging   `json:"logging"`
	Webhook  []Webhook `json:"webhook"`
	Health   Health    `json:"health"`
}

type Storage struct {
	Location              string `json:"location"`
	RedisHost             string `json:"redishost"`
	RedisPort             string `json:"redisport"`
	RedisPassword         string `json:"redispassword"`
	RedisConnectionString string `json:"redisconnectionstring"`
}

type Analytics struct {
	CollectionInterval int    `json:"collectioninterval"`
	RouteName          string `json:"routename"`
}

type Default struct {
	Mode                     string `json:"mode"`
	LogPath                  string `json:"logpath"`
	EnableDnsChallengeSolver bool   `json:"enablednschallengesolver"`
	DataPersistancePath      string `json:"datapersistancepath"`
	TIMEOUT                  struct {
		Connect string `json:"connect"`
		Server  string `json:"server"`
		Client  string `json:"client"`
	} `json:"timeout"`
	Analytics      Analytics `json:"analytics"`
	Storage        Storage   `json:"storage"`
	ErrorResponses struct {
		ErrorPageButtonName string `json:"errorpagebuttonname"`
		ErrorPageButtonUrl  string `json:"errorpagebuttonurl"`
	}
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
			MultipleCertAndKeyLocation string `json:"multiplecertandkeylocation"`
		} `json:"secure"`
	} `json:"bind"`
	SecureVerify    string   `json:"secureverify"`
	Secure          bool     `json:"secure"`
	Options         []string `json:"options"`
	DefaultBackend  string   `json:"defaultbackend"`
	FallbackBackend string   `json:"fallbackbackend"`
	Balance         string   `json:"balance"`
}

type Backend struct {
	Name                       string `json:"name"`
	HealthCheckMode            string `json:"healthcheckmode"`
	HealthCheckTriggerDuration int    `json:"healthchecktriggerduration"`
	Servers                    []struct {
		Id        string `json:"id"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		HealthUrl string `json:"healthurl"`
	}
}

type Logging struct {
	Enable      bool `json:"enable"`
	EnableRmote bool `json:"enableremote"`
	RemoteBind  struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"remotebind"`
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
