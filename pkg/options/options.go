package options

type Options struct {
	XmtpGrpcAddress    string `short:"x" long:"xmtp-address" env:"XMTP_GRPC_ADDRESS" description:"Address (including port) of XMTP GRPC server"`
	DbConnectionString string `short:"d" long:"db-connection-string" env:"DB_CONNECTION_STRING" description:"Address to database"`
	LogEncoding        string `long:"log-encoding" env:"LOG_ENCODING" description:"Log encoding" choice:"console" choice:"json" default:"console"`
	LogLevel           string `long:"log-level" env:"LOG_LEVEL" description:"log-level" choice:"debug" choice:"info" choice:"error" default:"info"`
	CreateMigration    string `long:"create-migration" description:"create a migration with the given name"`
}
