package v1

import (
	"context"

	"github.com/chekist32/go-monero/utils"
	"github.com/chekist32/goipay/internal/db"
	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGrpc struct {
	dbConnPool *pgxpool.Pool
	log        *zerolog.Logger
	pb_v1.UnimplementedUserServiceServer
}

func (u *UserGrpc) createUser(ctx context.Context, q *db.Queries, in *pb_v1.RegisterUserRequest) (*pgtype.UUID, error) {
	// Without userId in the request
	if in.UserId == nil {
		userId, err := q.CreateUser(ctx)
		if err != nil {
			u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "CreateUser").Msg(util.DefaultFailedSqlQueryMsg)
			return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}

		return &userId, err
	}

	// With userId in the request
	userIdReq, err := util.StringToPgUUID(*in.UserId)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg(util.InvalidUserIdInvalidUUIDMsg)
		return nil, status.Error(codes.InvalidArgument, util.InvalidUserIdInvalidUUIDMsg)
	}

	userId, err := q.CreateUserWithId(ctx, *userIdReq)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "CreateUserWithId").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	return &userId, err
}

func (u *UserGrpc) RegisterUser(ctx context.Context, in *pb_v1.RegisterUserRequest) (*pb_v1.RegisterUserResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, u.dbConnPool)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}
	defer tx.Rollback(ctx)

	userId, err := u.createUser(ctx, q, in)
	if err != nil {
		return nil, err
	}

	_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{UserID: *userId})
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "CreateUser").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	tx.Commit(ctx)

	return &pb_v1.RegisterUserResponse{UserId: util.PgUUIDToString(*userId)}, nil
}

func (u *UserGrpc) handleXmrCryptoDataUpdate(ctx context.Context, q *db.Queries, in *pb_v1.XmrKeysUpdateRequest, cryptData *db.CryptoDatum) error {
	_, err := utils.NewPrivateKey(in.PrivViewKey)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg("An error occurred while creating the XMR private view key.")
		return status.Error(codes.InvalidArgument, "Invalid XMR private view key.")
	}
	_, err = utils.NewPublicKey(in.PubSpendKey)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg("An error occurred while creating the XMR public spend key.")
		return status.Error(codes.InvalidArgument, "Invalid XMR public spend key.")
	}

	_, err = q.DeleteAllCryptoAddressByUserIdAndCoin(ctx, db.DeleteAllCryptoAddressByUserIdAndCoinParams{Coin: db.CoinTypeXMR, UserID: cryptData.UserID})
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "DeleteAllCryptoAddressByUserIdAndCoin").Msg(util.DefaultFailedSqlQueryMsg)
		return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	if !cryptData.XmrID.Valid {
		xmrData, err := q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: in.PrivViewKey, PubSpendKey: in.PubSpendKey})
		if err != nil {
			u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "CreateXMRCryptoData").Msg(util.DefaultFailedSqlQueryMsg)
			return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}
		_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: cryptData.UserID, XmrID: xmrData.ID})
		if err != nil {
			u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "SetXMRCryptoDataByUserId").Msg(util.DefaultFailedSqlQueryMsg)
			return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}
		return nil
	}
	_, err = q.UpdateKeysXMRCryptoDataById(ctx, db.UpdateKeysXMRCryptoDataByIdParams{ID: cryptData.XmrID, PrivViewKey: in.PrivViewKey, PubSpendKey: in.PubSpendKey})
	if err != nil {
		return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	return nil
}

func (u *UserGrpc) UpdateCryptoKeys(ctx context.Context, in *pb_v1.UpdateCryptoKeysRequest) (*pb_v1.UpdateCryptoKeysResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, u.dbConnPool)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}
	defer tx.Rollback(ctx)

	userId, err := util.StringToPgUUID(in.UserId)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg("An error occurred while converting the string to the PostgreSQL UUID data type.")
		return nil, status.Error(codes.InvalidArgument, util.InvalidUserIdInvalidUUIDMsg)
	}

	if err := checkIfUserExistsUUID(ctx, u.log, q, *userId); err != nil {
		return nil, err
	}

	cryptData, err := q.FindCryptoDataByUserId(ctx, *userId)
	if err != nil {
		u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Str("queryName", "FindCryptoDataByUserId").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	if in.XmrReq != nil {
		if err := u.handleXmrCryptoDataUpdate(ctx, q, in.XmrReq, &cryptData); err != nil {
			u.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg("")
			return nil, err
		}
	}

	tx.Commit(ctx)

	return &pb_v1.UpdateCryptoKeysResponse{}, nil
}

func NewUserGrpc(dbConnPool *pgxpool.Pool, log *zerolog.Logger) *UserGrpc {
	return &UserGrpc{dbConnPool: dbConnPool, log: log}
}
