package app

import (
	"context"

	"github.com/chekist32/goipay/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RequestLoggingInterceptor struct {
	log *zerolog.Logger
}

func (i *RequestLoggingInterceptor) Intercepte(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	i.log.Info().Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msgf("PRE %s", info.FullMethod)

	res, err := handler(ctx, req)
	if err != nil {
		i.log.Info().Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("status", "failure").Msgf("POST %s", info.FullMethod)
	} else {
		i.log.Info().Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("status", "success").Msgf("POST %s", info.FullMethod)
	}

	return res, err
}

func NewRequestLoggingInterceptor(log *zerolog.Logger) *RequestLoggingInterceptor {
	return &RequestLoggingInterceptor{log: log}
}

type MetadataInterceptor struct {
	log *zerolog.Logger
}

func (i *MetadataInterceptor) Intercepte(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	i.log.Debug().Msg("PRE MetadataInterceptor")

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		i.log.Debug().Msg("Failed to obtain metadata")
		return handler(ctx, req)
	}

	getReqIdOrCreate := func() string {
		reqIdSlice := md[util.RequestIdKey]
		if len(reqIdSlice) < 1 || reqIdSlice[0] == "" {
			reqId := uuid.NewString()
			i.log.Debug().Msgf("Generating a new requestId: %v", reqId)
			return reqId
		}

		return reqIdSlice[0]
	}

	metadataCtx := context.WithValue(ctx, util.MetadataCtxKey, util.CustomMetadata{RequestId: getReqIdOrCreate()})

	i.log.Debug().Msg("POST MetadataInterceptor")

	return handler(metadataCtx, req)

}

func NewMetadataInterceptor(log *zerolog.Logger) *MetadataInterceptor {
	return &MetadataInterceptor{log: log}
}
