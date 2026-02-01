package config

type Postgres struct {
	Host string `env:"POSTGRES_HOST" koanf:"host"`
	Port string `env:"POSTGRES_PORT" koanf:"port"`
	Name string `env:"POSTGRES_DATABASE_NAME" koanf:"db_name"`
	User string `env:"POSTGRES_USER" koanf:"user"`
	Pass string `env:"POSTGRES_PASS" koanf:"pass"`
	SSL  bool   `env:"POSTGRES_SSL" envDefault:"false" koanf:"ssl"`
}

type HTTP struct {
	Port      string `env:"HTTP_PORT" envDefault:"9090" koanf:"port"`
	AdminPass string `env:"ADMIN_PASS" koanf:"admin_pass"`
}

type GoEnv struct {
	Mode string `env:"GO_ENV" envDefault:"debug" koanf:"mode"`
}

type Notifications struct {
	Asiatech struct {
		Host     string `env:"ASIATECH_HOST" koanf:"host"`
		Username string `env:"ASIATECH_USERNAME" koanf:"username"`
		Password string `env:"ASIATECH_PASSWORD" koanf:"password"`
		Scope    string `env:"ASIATECH_SCOPE" koanf:"scope"`
		Sender   string `env:"ASIATECH_SENDER" koanf:"sender"`
		Priority int    `env:"ASIATECH_PRIORITY" envDefault:"4" koanf:"priority"`
		Enabled  bool   `env:"ASIATECH_ENABLED" envDefault:"false" koanf:"enabled"`
	} `knoanf:"asiatech"`
	Smsir struct {
		ApiKey     string `env:"SMSIR_API_TOKEN" koanf:"api_key"`
		LineNumber string `env:"SMSIR_LINE_NUMBER" koanf:"line_number"`
		Enabled    bool   `env:"SMSIR_ENABLED" envDefault:"false" koanf:"enabled"`
		Priority   int    `env:"SMSIR_PRIORITY" envDefault:"2" koanf:"priority"`
	} `koanf:"smsir"`
	Kavenegar struct {
		ApiToken string `env:"KAVENEGAR_API_TOKEN" koanf:"api_token"`
		Sender   string `env:"KAVENEGAR_SENDER" envDefault:"" koanf:"sender"`
		Enabled  bool   `env:"KAVENEGAR_ENABLED" envDefault:"true" koanf:"enabled"`
		Priority int    `env:"KAVENEGAR_PRIORITY" envDefault:"1" koanf:"priority"`
	} `koanf:"kavenegar"`
	Email struct {
		Host     string `env:"EMAIL_HOST" koanf:"host"`
		Port     string `env:"EMAIL_PORT" koanf:"port"`
		User     string `env:"EMAIL_USER" koanf:"user"`
		Password string `env:"EMAIL_PASSWORD" koanf:"password"`
		From     string `env:"EMAIL_FROM" koanf:"from"`
		Enabled  bool   `env:"EMAIL_ENABLED" envDefault:"false" koanf:"enabled"`
	} `koanf:"email"`
	Telegram struct {
		BotToken string `env:"TELEGRAM_BOT_TOKEN" koanf:"bot_token"`
		Proxy    string `env:"TELEGRAM_PROXY" envDefault:"" koanf:"proxy"`
		Enabled  bool   `env:"TELEGRAM_ENABLED" envDefault:"false" koanf:"enabled"`
	} `env:"TELEGRAM_ENABLED" envDefault:"false" koanf:"telegram"`
	Mail struct {
		SMTPHost    string `env:"MAIL_SMTP_HOST" koanf:"smtp_host"`
		SMTPPort    int    `env:"MAIL_SMTP_PORT" koanf:"smtp_port"`
		Username    string `env:"MAIL_USERNAME" koanf:"username"`
		Password    string `env:"MAIL_PASSWORD" koanf:"password"`
		FromAddress string `env:"MAIL_FROM_ADDRESS" koanf:"from_address"`
		FromName    string `env:"MAIL_FROM_NAME" koanf:"from_name"`
		Enabled     bool   `env:"MAIL_ENABLED" envDefault:"false" koanf:"enabled"`
		Priority    int    `env:"MAIL_PRIORITY" envDefault:"5" koanf:"priority"`
	} `koanf:"mail"`
	Mattermost struct {
		Url      string `env:"MATTERMOST_URL" koanf:"url"`
		BotToken string `env:"MATTERMOST_BOT_TOKEN" koanf:"bot_token"`
		Enabled  bool   `env:"MATTERMOST_ENABLED" envDefault:"false" koanf:"enabled"`
		Priority int    `env:"MATTERMOST_PRIORITY" envDefault:"3" koanf:"priority"`
	} `koanf:"mattermost"`
}

type Scheduler struct {
	MobileScheduler struct {
		StartAt       string `env:"MOBILE_SCHEDULER_START_AT" envDefault:"1s" koanf:"start_at"`
		Interval      string `env:"MOBILE_SCHEDULER_INTERVAL" envDefault:"600s" koanf:"interval"`
		Workers       int    `env:"MOBILE_SCHEDULER_WORKERS" envDefault:"1" koanf:"workers"`
		QueueSize     int    `env:"MOBILE_SCHEDULER_QUEUE_SIZE" envDefault:"1" koanf:"queue_size"`
		CacheCapacity int    `env:"MOBILE_SCHEDULER_CACHE_CAPACITY" envDefault:"10" koanf:"cache_capacity"`
	} `koanf:"mobile_scheduler"`
	AlertScheduler struct {
		StartAt   string `env:"ALERT_SCHEDULER_START_AT" envDefault:"3s" koanf:"start_at"`
		Interval  string `env:"ALERT_SCHEDULER_INTERVAL" envDefault:"10s" koanf:"interval"`
		Workers   int    `env:"ALERT_SCHEDULER_WORKERS" envDefault:"1" koanf:"workers"`
		QueueSize int    `env:"ALERT_SCHEDULER_QUEUE_SIZE" envDefault:"10" koanf:"queue_size"`
	} `koanf:"alert_scheduler"`
	MessageStatus struct {
		StartAt   string `env:"MESSAGE_STATUS_START_AT" envDefault:"4s" koanf:"start_at"`
		Interval  string `env:"MESSAGE_STATUS_INTERVAL" envDefault:"20s" koanf:"interval"`
		Workers   int    `env:"MESSAGE_STATUS_WORKERS" envDefault:"10" koanf:"workers"`
		QueueSize int    `env:"MESSAGE_STATUS_QUEUE_SIZE" envDefault:"100" koanf:"queue_size"`
	} `koanf:"message_status"`
	Enabled bool `env:"SCHEDULER_ENABLED" envDefault:"false" koanf:"scheduler_enabled"`
}

type Config struct {
	Postgres      Postgres      `koanf:"postgres"`
	HTTP          HTTP          `koanf:"http"`
	Go            GoEnv         `koanf:"go"`
	Notifications Notifications `koanf:"notifications"`
	Scheduler     Scheduler     `koanf:"scheduler"`
	JwtSecret     string        `env:"JWT_SECRET" koanf:"jwt_secret"`
	SignupEnabled bool          `env:"SIGNUP_ENABLED" envDefault:"true" koanf:"signup_enabled"`
}
