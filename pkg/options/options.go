package options

type ApiOptions struct {
	Enabled bool `long:"api" description:"Enable the GRPC API server"`
	Port    int  `short:"p" long:"api-port" env:"API_PORT" default:"8080" description:"Port for the Connect GRPC API"`
}

type ApnsOptions struct {
	Enabled               bool   `long:"apns-enabled" env:"APNS_ENABLED" description:"Enable APNS"`
	P8Certificate         string `long:"apns-p8-certificate" env:"APNS_P8_CERTIFICATE" description:".p8 certificate data for APNS"`
	P8CertificateFilePath string `long:"apns-p8-certificate-file-path" env:"APNS_P8_CERTIFICATE_FILE_PATH" description:".p8 certificate file for APNS"`
	KeyId                 string `long:"apns-key-id" env:"APNS_KEY_ID" description:"Key ID associated with APNS credentials"`
	TeamId                string `long:"apns-team-id" env:"APNS_TEAM_ID" description:"APNS Team ID"`
	Topic                 string `long:"apns-topic" env:"APNS_TOPIC" description:"Topic to be used on all messages"`
	Mode                  string `long:"apns-mode" env:"APNS_MODE" description:"Which APNS servers to deliver to, development or production" choice:"development" choice:"production" default:"development"`
}

type FcmOptions struct {
	Enabled         bool   `long:"fcm-enabled" env:"FCM_ENABLED" description:"Enable FCM sending"`
	CredentialsJson string `long:"fcm-credentials-json" env:"FCM_CREDENTIALS_JSON" description:"FCM Credentials"`
	ProjectId       string `long:"fcm-project-id" env:"FCM_PROJECT_ID" description:"FCM Project ID"`
}

type XmtpOptions struct {
	ListenerEnabled bool   `long:"xmtp-listener" description:"Enable the XMTP listener to actually send notifications. Requires APNSOptions to be configured"`
	UseTls          bool   `long:"xmtp-listener-tls" description:"Whether to connect to XMTP network using TLS"`
	GrpcAddress     string `short:"x" long:"xmtp-address" env:"XMTP_GRPC_ADDRESS" description:"Address (including port) of XMTP GRPC server"`
	NumWorkers      int    `long:"num-workers" description:"Number of workers used to process messages" default:"50"`
}

type HttpDeliveryOptions struct {
	Enabled    bool   `long:"http-delivery"`
	Address    string `long:"http-delivery-address"`
	AuthHeader string `long:"http-auth-header"`
}

type Options struct {
	Api          ApiOptions          `group:"API Options"`
	Xmtp         XmtpOptions         `group:"Worker Options"`
	Apns         ApnsOptions         `group:"APNS Options"`
	Fcm          FcmOptions          `group:"FCM Options"`
	HttpDelivery HttpDeliveryOptions `group:"HTTP Delivery Options"`

	HsEnv		   string `short:"e" long:"env" env:"HS_ENV" description:"Deployment environment"`
	DbConnectionString string `short:"d" long:"db-connection-string" env:"DB_CONNECTION_STRING" description:"Address to database"`
	LogEncoding        string `long:"log-encoding" env:"LOG_ENCODING" description:"Log encoding" choice:"console" choice:"json" default:"console"`
	LogLevel           string `long:"log-level" env:"LOG_LEVEL" description:"log-level" choice:"debug" choice:"info" choice:"error" default:"info"`
	CreateMigration    string `long:"create-migration" description:"create a migration with the given name"`
}
