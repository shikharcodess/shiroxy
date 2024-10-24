package models

type Config struct {
	Environment Environment `json:"environment"`
	Default     Default     `json:"default"`
	Frontend    Frontend    `json:"frontend"`
	Backend     Backend     `json:"backend"`
	Logging     Logging     `json:"logging"`
	Webhook     Webhook     `json:"webhook"`
	Health      Health      `json:"health"`
}

type Environment struct {
	Mode                         string `json:"mode"`
	InstanceName                 string `json:"instancename"`
	AcmeServerUrl                string `json:"acmeserverurl"`
	ACMEServerInsecureSkipVerify string `json:"aCMEserverinsecureskipverify"`
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
	DebugMode                string       `json:"debugmode"`
	LogPath                  string       `json:"logpath"`
	EnableDnsChallengeSolver bool         `json:"enablednschallengesolver"`
	DataPersistancePath      string       `json:"datapersistancepath"`
	Analytics                Analytics    `json:"analytics"`
	Storage                  Storage      `json:"storage"`
	ErrorResponses           ErrorRespons `json:"errorresponses"`
	TIMEOUT                  struct {
		Connect string `json:"connect"`
		Server  string `json:"server"`
		Client  string `json:"client"`
	} `json:"timeout"`

	User struct {
		Email  string `json:"email"`
		Secret string `json:"secret"`
	} `json:"user"`
	AdminAPI struct {
		Port string `json:"port"`
	} `json:"adminapi"`
}

type ErrorRespons struct {
	ErrorPageButtonName string `json:"errorpagebuttonname"`
	ErrorPageButtonUrl  string `json:"errorpagebuttonurl"`
}

type FrontendBind struct {
	Port          string                  `json:"port"`
	Host          string                  `json:"host"`
	Target        string                  `json:"target"`
	Secure        bool                    `json:"secure"`
	SecureSetting FrontendSecuritySetting `json:"securesetting"`
}

type FrontendSecuritySetting struct {
	// required, optional, none
	SecureVerify string `json:"secureverify"`
	// certandkey and shiroxyshinglesecure
	SingleTargetMode    string                                     `json:"singletargetmode"`
	CertAndKey          FrontendSecuritySettingCertAndKey          `json:"certandkey"`
	ShiroxySingleSecure FrontendSecuritySettingShiroxySingleSecure `json:"shiroxysinglesecure"`
}

type FrontendSecuritySettingCertAndKey struct {
	Domain string `json:"domain"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}

type FrontendSecuritySettingShiroxySingleSecure struct {
	Domain string `json:"domain"`
}

type Frontend struct {
	Mode            string         `json:"mode"`
	HttpToHttps     bool           `json:"httptohttps"`
	Bind            []FrontendBind `json:"bind"`
	Options         []string       `json:"options"`
	DefaultBackend  string         `json:"defaultbackend"`
	FallbackBackend string         `json:"fallbackbackend"`
}

type Backend struct {
	Name                       string `json:"name"`
	Balance                    string `json:"balance"`
	HealthCheckMode            string `json:"healthcheckmode"`
	HealthCheckTriggerDuration int    `json:"healthchecktriggerduration"`
	Tagrule                    string `json:"tagrule"`
	NoServerAction             string `json:"noserveraction"`
	Servers                    []BackendServer
	Tags                       []string `json:"tags"`
}

type BackendServer struct {
	Id        string `json:"id"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	HealthUrl string `json:"healthurl"`
	Tags      string `json:"tags"`
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
