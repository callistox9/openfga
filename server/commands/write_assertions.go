package commands

import (
	"context"

	"github.com/openfga/openfga/pkg/logger"
	tupleUtils "github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/utils"
	serverErrors "github.com/openfga/openfga/server/errors"
	"github.com/openfga/openfga/storage"
	openfgapb "go.buf.build/openfga/go/openfga/api/openfga/v1"
)

type WriteAssertionsCommand struct {
	assertionBackend      storage.AssertionsBackend
	typeDefinitionBackend storage.TypeDefinitionReadBackend
	logger                logger.Logger
}

func NewWriteAssertionsCommand(
	assertionBackend storage.AssertionsBackend,
	typeDefinitionBackend storage.TypeDefinitionReadBackend,
	logger logger.Logger,
) *WriteAssertionsCommand {
	return &WriteAssertionsCommand{
		assertionBackend:      assertionBackend,
		typeDefinitionBackend: typeDefinitionBackend,
		logger:                logger,
	}
}

func (w *WriteAssertionsCommand) Execute(ctx context.Context, req *openfgapb.WriteAssertionsRequest) (*openfgapb.WriteAssertionsResponse, error) {
	store := req.GetStoreId()
	modelID := req.GetAuthorizationModelId()
	assertions := req.GetAssertions()

	dbCallsCounter := utils.NewDBCallCounter()
	for _, assertion := range assertions {
		if _, err := tupleUtils.ValidateTuple(ctx, w.typeDefinitionBackend, store, modelID, assertion.TupleKey, dbCallsCounter); err != nil {
			return nil, serverErrors.HandleTupleValidateError(err)
		}
	}
	dbCallsCounter.AddWriteCall()
	utils.LogDBStats(ctx, w.logger, "WriteAssertions", dbCallsCounter.GetReadCalls(), dbCallsCounter.GetWriteCalls())

	err := w.assertionBackend.WriteAssertions(ctx, store, modelID, assertions)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	return &openfgapb.WriteAssertionsResponse{}, nil
}
