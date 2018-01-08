package adapters

import (
	"github.com/smartcontractkit/chainlink-go/store"
	"github.com/smartcontractkit/chainlink-go/store/models"
)

type NoOp struct{}

func (self *NoOp) Perform(input models.RunResult, _ *store.Store) models.RunResult {
	return models.RunResult{}
}

type NoOpPend struct{}

func (self *NoOpPend) Perform(input models.RunResult, _ *store.Store) models.RunResult {
	return models.RunResultPending(input)
}