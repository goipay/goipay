package util

import "errors"

type contextKey string

const (
	DefaultFailedSqlTxInitMsg                    string = "An error occurred while initiating an SQL transaction."
	DefaultFailedSqlQueryMsg                     string = "An error occurred while executing a SQL query."
	DefaultFailedScanningToPostgresqlDataTypeMsg string = "An error occurred while scanning the value into a PostgreSQL data type."
	DefaultFailedFetchingXMRDaemonMsg            string = "An error occurred while fetching."

	InvalidUserIdInvalidUUIDMsg string = "Invalid userId (invalid UUID)."
	InvalidUserIdUserExistsMsg  string = "Invalid userId (user exists)."
)

const (
	RequestIdKey              = "request-id"
	MetadataCtxKey contextKey = "metadata"
)

const (
	RequestIdLogKey = "requestId"
)

var (
	invalidProtoBufCoinTypeErr error = errors.New("invalid protoBuf coin type")
	invalidDbCoinTypeErr       error = errors.New("invalid db coin type")
	invalidDbStatusTypeErr     error = errors.New("invalid db status type")
)
