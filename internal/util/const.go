package util

import (
	"errors"
	"time"
)

type contextKey string

const (
	MIN_SYNC_TIMEOUT     time.Duration = 10 * time.Second
	SEND_TIMEOUT         time.Duration = 10 * time.Second
	HEALTH_CHECK_TIEMOUT time.Duration = 5 * time.Second
)

const (
	DefaultFailedSqlTxInitMsg                    string = "An error occurred while initiating an SQL transaction."
	DefaultFailedSqlQueryMsg                     string = "An error occurred while executing a SQL query."
	DefaultFailedScanningToPostgresqlDataTypeMsg string = "An error occurred while scanning the value into a PostgreSQL data type."
	DefaultFailedFetchingDaemonMsg               string = "An error occurred while fetching."

	FailedStringToPgUUIDMappingMsg string = "An error occurred while converting the string to the PostgreSQL UUID data type."

	InvalidUserIdInvalidUUIDMsg      string = "Invalid userId (invalid UUID)."
	InvalidUserIdUserExistsMsg       string = "Invalid userId (user exists)."
	InvalidUserIdUserDoesNotExistMsg string = "Invalid userId (user does not exist)."

	InvoiceAmountBelow0ErrorMsg      string = "Invoice amount can't be below 0."
	InvoiceErrorWhileHandlingMsg     string = "An error occurred while handling invoice."
	InvoiceStreamSendingDataErrorMsg string = "An error occurred while sending data."
	InvoiceStreamClosedErrorMsg      string = "Stream has been closed."
)

const (
	RequestIdKey   string     = "request-id"
	MetadataCtxKey contextKey = "metadata"
)

const (
	RequestIdLogKey = "requestId"
)

var (
	invalidProtoBufCoinTypeErr error = errors.New("invalid protoBuf coin type")
	invalidDbCoinTypeErr       error = errors.New("invalid db coin type")
	invalidDbStatusTypeErr     error = errors.New("invalid db status type")

	InvalidNetworkTypeErr error = errors.New("invalid network type")
)
