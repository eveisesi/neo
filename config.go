package neo

type Config struct {
	// db configuration
	DBUser         string `required:"true"`
	DBPass         string `required:"true"`
	DBHost         string `required:"true"`
	DBName         string `required:"true"`
	DBReadTimeout  int    `default:"30"`
	DBWriteTimeout int    `default:"30"`

	// logger configuration
	LogLevel string `required:"true"`

	// ESI configuration
	ESIHost   string `required:"true"`
	ESIUAgent string `required:"true"`

	// redis configuration
	RedisAddr string `required:"true"`

	// zkillboard params
	ZUAgent string `required:"true"`

	ServerPort uint `envconfig:"SERVER_PORT" required:"true"`

	SSOClientID         string `envconfig:"SSO_CLIENT_ID" required:"true"`
	SSOClientSecret     string `envconfig:"SSO_CLIENT_SECRET" required:"true"`
	SSOCallback         string `envconfig:"SSO_CALLBACK" required:"true"`
	SSOAuthorizationURL string `envconfig:"SSO_AUTHORIZATION_URL" required:"true"`
	SSOTokenURL         string `envconfig:"SSO_TOKEN_URL" required:"true"`
	SSOJWKSURL          string `envconfig:"SSO_JWKS_URL" required:"true"`

	SlackNotifierEnabled        bool   `envconfig:"SLACK_NOTIFIER_ENABLED" default:"false"`
	SlackNotifierWebhookURL     string `envconfig:"SLACK_NOTIFIER_URL"`
	SlackNotifierValueThreshold int    `envconfig:"SLACK_NOTIFIER_THRESHOLD"`
	SlackActionBaseURL          string `envconfig:"SLACK_ACTION_BASE_URL"`

	SpacesEnabled  bool   `envconfig:"SPACES_ENABLED" required:"true"`
	SpacesBucket   string `envconfig:"SPACES_BUCKET"`
	SpacesEndpoint string `envconfig:"SPACES_ENDPOINT"`
	SpacesRegion   string `envconfig:"SPACES_REGION"`
	SpacesKey      string `envconfig:"SPACES_KEY"`
	SpacesSecret   string `envconfig:"SPACES_SECRET"`
}
