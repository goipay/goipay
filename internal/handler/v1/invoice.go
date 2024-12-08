package v1

import (
	"context"

	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/chekist32/goipay/internal/processor"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceGrpc struct {
	dbConnPool       *pgxpool.Pool
	log              *zerolog.Logger
	paymentProcessor *processor.PaymentProcessor
	pb_v1.UnimplementedInvoiceServiceServer
}

func (i *InvoiceGrpc) CreateInvoice(ctx context.Context, req *pb_v1.CreateInvoiceRequest) (*pb_v1.CreateInvoiceResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, i.dbConnPool)
	if err != nil {
		i.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}
	defer tx.Rollback(ctx)

	if req.Amount < 0 {
		return nil, status.Error(codes.InvalidArgument, "Invoice amount can't be below 0.")
	}
	if err := checkIfUserExistsString(ctx, i.log, q, req.UserId); err != nil {
		return nil, err
	}

	invoice, err := i.paymentProcessor.HandleNewInvoice(util.PbNewInvoiceToProcessorNewInvoice(req))
	if err != nil {
		errMsg := "An error occurred while handling invoice."
		i.log.Err(err).Str(util.RequestIdLogKey, util.GetRequestIdOrEmptyString(ctx)).Msg(errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	tx.Commit(ctx)

	return &pb_v1.CreateInvoiceResponse{PaymentId: util.PgUUIDToString(invoice.ID), Address: invoice.CryptoAddress}, nil
}

func (i *InvoiceGrpc) InvoiceStatusStream(req *pb_v1.InvoiceStatusStreamRequest, stream pb_v1.InvoiceService_InvoiceStatusStreamServer) error {
	invoiceCn := i.paymentProcessor.NewInvoicesChan()

	for {
		select {
		case invoice := <-invoiceCn:
			if err := stream.Send(&pb_v1.InvoiceStatusStreamResponse{Invoice: util.DbInvoiceToPbInvoice(&invoice)}); err != nil {
				errMsg := "An error occured while sending data."
				i.log.Err(err).Msg(errMsg)
				return status.Error(codes.Canceled, errMsg)
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream has been closed.")
		}
	}

}

func NewInvoiceGrpc(dbConnPool *pgxpool.Pool, paymentProcessor *processor.PaymentProcessor, log *zerolog.Logger) *InvoiceGrpc {
	return &InvoiceGrpc{dbConnPool: dbConnPool, paymentProcessor: paymentProcessor, log: log}
}
