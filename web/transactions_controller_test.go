package web_test

import (
	"math/big"
	"testing"

	"github.com/manyminds/api2go/jsonapi"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/store/presenters"
	"github.com/smartcontractkit/chainlink/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionsController_Index_Success(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	ethMock := app.MockEthClient()
	ethMock.Context("app.Start()", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_getTransactionCount", "0x100")
	})

	require.NoError(t, app.Start())
	store := app.GetStore()
	client := app.NewHTTPClient()

	from := cltest.GetAccountAddress(store)
	tx1 := cltest.CreateTxAndAttempt(store, from, 1)
	_, err := store.AddTxAttempt(tx1, tx1.EthTx(big.NewInt(2)), 2)
	require.NoError(t, err)
	tx2 := cltest.CreateTxAndAttempt(store, from, 3)
	tx3 := cltest.CreateTxAndAttempt(store, from, 4)
	_, count, err := store.Transactions(0, 100)
	require.NoError(t, err)
	require.Equal(t, count, 3)

	resp, cleanup := client.Get("/v2/transactions?size=2")
	defer cleanup()
	cltest.AssertServerResponse(t, resp, 200)

	var links jsonapi.Links
	var txs []presenters.Tx
	body := cltest.ParseResponseBody(resp)
	require.NoError(t, web.ParsePaginatedResponse(body, &txs, &links))
	assert.NotEmpty(t, links["next"].Href)
	assert.Empty(t, links["prev"].Href)

	require.Len(t, txs, 2)
	require.Equal(t, string(tx3.SentAt), txs[0].SentAt, "expected tx attempts order by sentAt descending")
	require.Equal(t, string(tx2.SentAt), txs[1].SentAt, "expected tx attempts order by sentAt descending")
}

func TestTransactionsController_Index_Error(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", "0x100")
	require.NoError(t, app.Start())

	client := app.NewHTTPClient()
	resp, cleanup := client.Get("/v2/transactions?size=TrainingDay")
	defer cleanup()
	cltest.AssertServerResponse(t, resp, 422)
}

func TestTransactionsController_Show_Success(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	ethMock := app.MockEthClient()
	ethMock.Context("app.Start()", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_getTransactionCount", "0x100")
	})

	require.NoError(t, app.Start())
	store := app.GetStore()
	client := app.NewHTTPClient()
	from := cltest.GetAccountAddress(store)
	tx := cltest.CreateTxAndAttempt(store, from, 1)
	ta1 := tx.TxAttempt
	ta2, err := store.AddTxAttempt(tx, tx.EthTx(big.NewInt(2)), 2)
	require.NoError(t, err)
	txWithAttempt1 := *tx
	txWithAttempt1.TxAttempt = ta1
	txWithAttempt2 := *tx
	txWithAttempt2.TxAttempt = *ta2

	tests := []struct {
		name string
		hash string
		want models.Tx
	}{
		{"old hash", ta1.Hash.String(), txWithAttempt1},
		{"current hash", ta2.Hash.String(), txWithAttempt2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, cleanup := client.Get("/v2/transactions/" + test.hash)
			defer cleanup()
			cltest.AssertServerResponse(t, resp, 200)

			ptx := presenters.Tx{}
			require.NoError(t, cltest.ParseJSONAPIResponse(resp, &ptx))

			test.want.ID = 0
			test.want.TxID = 0
			assert.Equal(t, presenters.NewTx(&test.want), ptx)
		})
	}
}

func TestTransactionsController_Show_NotFound(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	ethMock := app.MockEthClient()
	ethMock.Context("app.Start()", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_getTransactionCount", "0x100")
	})

	require.NoError(t, app.Start())
	store := app.GetStore()
	client := app.NewHTTPClient()
	from := cltest.GetAccountAddress(store)
	tx := cltest.CreateTxAndAttempt(store, from, 1)

	resp, cleanup := client.Get("/v2/transactions/" + (tx.Hash.String() + "1"))
	defer cleanup()
	cltest.AssertServerResponse(t, resp, 404)
}
