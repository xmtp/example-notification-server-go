package options

type ApiOptions struct {
	Enabled bool `long:"api" description:"Enable the GRPC API server"`
	Port    int  `short:"p" long:"api-port" env:"API_PORT" default:"8080" description:"Port for the Connect GRPC API"`
}

type WorkerOptions struct {
	Enabled    bool `long:"worker" description:"Enable the stream listening worker"`
	NumWorkers int  `long:"num-workers" description:"Number of workers used to process messages" default:"50"`
}

type Options struct {
	Api                ApiOptions    `group:"API Options"`
	Worker             WorkerOptions `group:"Worker Options"`
	XmtpGrpcAddress    string        `short:"x" long:"xmtp-address" env:"XMTP_GRPC_ADDRESS" description:"Address (including port) of XMTP GRPC server"`
	DbConnectionString string        `short:"d" long:"db-connection-string" env:"DB_CONNECTION_STRING" description:"Address to database"`
	LogEncoding        string        `long:"log-encoding" env:"LOG_ENCODING" description:"Log encoding" choice:"console" choice:"json" default:"console"`
	LogLevel           string        `long:"log-level" env:"LOG_LEVEL" description:"log-level" choice:"debug" choice:"info" choice:"error" default:"info"`
	CreateMigration    string        `long:"create-migration" description:"create a migration with the given name"`
}
