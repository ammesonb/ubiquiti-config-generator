package configuration

type Config struct {
	Logging     LoggingConfig `yaml:"logging"`
	Git         GitConfig     `yaml:"git"`
	DevicesFile string        `yaml:"devices-file"`
	Devices     []*DeviceConfig
}

type LoggingConfig struct {
	DBName   string `yaml:"dbConnection"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type GitConfig struct {
	AppID          string `yaml:"app-id"`
	PrimaryBranch  string `yaml:"primary-branch"`
	PrivateKeyPath string `yaml:"private-key-path"`
	WebhookURL     string `yaml:"webhook-url"`
	ListenIP       string `yaml:"listen-ip"`
	WebhookPort    string `yaml:"webhook-port"`
	WebhookSecret  string `yaml:"webhook-secret"`
}

type DeviceConfig struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	KeyFile  string `yaml:"keyfile"`

	TemplatesDir string   `yaml:"templatesDir"`
	ConfigFiles  []string `yaml:"configFiles"`

	// For generated NAT rules, what number to start with and steps between them
	NatRuleNumberStart int `yaml:"nat-rule-number-start"`
	NatRuleNumberStep  int `yaml:"nat-rule-number-step"`

	CommandFilePath     string `yaml:"command-file-path"`
	ConfigureScriptPath string `yaml:"configure-script-path"`

	RebootAfterMinutes int32 `yaml:"reboot-after-minutes"`
	AutoRollBack       bool  `yaml:"auto-rollback-on-failure"`
	SaveAfterCommit    bool  `yaml:"save-after-commit"`
}
