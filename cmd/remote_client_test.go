package cmd_test

import (
	"flag"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/cmd"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/store/presenters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestClient_DisplayAccountBalance(t *testing.T) {
	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getBalance", "0x0100")
	ethMock.Register("eth_call", "0x0100")

	client, r := app.NewClientAndRenderer()

	assert.Nil(t, client.DisplayAccountBalance(cltest.EmptyCLIContext()))
	require.Equal(t, 1, len(r.Renders))
	from := cltest.GetAccountAddress(app.GetStore())
	balances := *r.Renders[0].(*[]presenters.AccountBalance)
	assert.Equal(t, from.Hex(), balances[0].Address)
}

func TestClient_GetJobSpecs(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()

	j1 := cltest.NewJob()
	app.Store.SaveJob(&j1)
	j2 := cltest.NewJob()
	app.Store.SaveJob(&j2)

	client, r := app.NewClientAndRenderer()

	require.Nil(t, client.GetJobSpecs(cltest.EmptyCLIContext()))
	jobs := *r.Renders[0].(*[]models.JobSpec)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, j1.ID, jobs[0].ID)
}

func TestClient_ShowJobRun_Exists(t *testing.T) {
	t.Parallel()
	app, cleanup := cltest.NewApplication()
	defer cleanup()

	j, _ := cltest.NewJobWithWebInitiator()
	assert.NoError(t, app.Store.SaveJob(&j))

	jr := cltest.CreateJobRunViaWeb(t, app, j, `{"value":"100"}`)

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{jr.ID})
	c := cli.NewContext(nil, set, nil)
	assert.NoError(t, client.ShowJobRun(c))
	assert.Equal(t, 1, len(r.Renders))
	assert.Equal(t, jr.ID, r.Renders[0].(*presenters.JobRun).ID)
}

func TestClient_ShowJobRun_NotFound(t *testing.T) {
	t.Parallel()
	app, cleanup := cltest.NewApplication()
	defer cleanup()

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"bogus-ID"})
	c := cli.NewContext(nil, set, nil)
	assert.Error(t, client.ShowJobRun(c))
	assert.Empty(t, r.Renders)
}

func TestClient_ShowJobSpec_Exists(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	job := cltest.NewJob()
	app.Store.SaveJob(&job)

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{job.ID})
	c := cli.NewContext(nil, set, nil)
	require.Nil(t, client.ShowJobSpec(c))
	require.Equal(t, 1, len(r.Renders))
	assert.Equal(t, job.ID, r.Renders[0].(*presenters.JobSpec).ID)
}

func TestClient_ShowJobSpec_NotFound(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"bogus-ID"})
	c := cli.NewContext(nil, set, nil)
	assert.Error(t, client.ShowJobSpec(c))
	assert.Empty(t, r.Renders)
}

func TestClient_CreateServiceAgreement(t *testing.T) {
	config, _ := cltest.NewConfigWithPrivateKey()
	app, cleanup := cltest.NewApplicationWithConfigAndUnlockedAccount(config)
	defer cleanup()
	client, _ := app.NewClientAndRenderer()

	sa := cltest.EasyJSONFromFixture("../internal/fixtures/web/hello_world_agreement.json")

	tests := []struct {
		name        string
		input       string
		jobsCreated bool
		errored     bool
	}{
		{"invalid json", "{bad son}", false, true},
		{"bad file path", "bad/filepath/", false, true},
		{"valid service agreement", sa.String(), true, false},
		{"service agreement specified as path", "../internal/fixtures/web/hello_world_agreement.json", true, false},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {

			set := flag.NewFlagSet("create", 0)
			assert.NoError(t, set.Parse([]string{test.input}))
			c := cli.NewContext(nil, set, nil)

			err := client.CreateServiceAgreement(c)

			cltest.AssertError(t, test.errored, err)
			jobs := cltest.AllJobs(app.Store)
			if test.jobsCreated {
				assert.True(t, len(jobs) > 0)
			} else {
				assert.Equal(t, 0, len(jobs))
			}
		})
	}
}

func TestClient_CreateJobSpec(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	client, _ := app.NewClientAndRenderer()

	tests := []struct {
		name, input string
		nJobs       int
		errored     bool
	}{
		{"bad json", "{bad son}", 0, true},
		{"bad filepath", "bad/filepath/", 0, true},
		{"web", `{"initiators":[{"type":"web"}],"tasks":[{"type":"NoOp"}]}`, 1, false},
		{"runAt", `{"initiators":[{"type":"runAt","params":{"time":"2018-01-08T18:12:01.103Z"}}],"tasks":[{"type":"NoOp"}]}`, 2, false},
		{"file", "../internal/fixtures/web/end_at_job.json", 3, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set := flag.NewFlagSet("create", 0)
			set.Parse([]string{test.input})
			c := cli.NewContext(nil, set, nil)

			err := client.CreateJobSpec(c)
			cltest.AssertError(t, test.errored, err)

			numberOfJobs := cltest.AllJobs(app.Store)
			assert.Equal(t, test.nJobs, len(numberOfJobs))
		})
	}
}

func TestClient_CreateJobSpec_JSONAPIErrors(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	client, _ := app.NewClientAndRenderer()

	set := flag.NewFlagSet("create", 0)
	set.Parse([]string{`{"initiators":[{"type":"runAt"}],"tasks":[{"type":"NoOp"}]}`})
	c := cli.NewContext(nil, set, nil)

	err := client.CreateJobSpec(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have a time")
}

func TestClient_CreateJobRun(t *testing.T) {
	t.Parallel()
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	client, _ := app.NewClientAndRenderer()

	tests := []struct {
		name    string
		json    string
		jobSpec models.JobSpec
		errored bool
	}{
		{"CreateSuccess", `{"value": 100}`, first(cltest.NewJobWithWebInitiator()), false},
		{"EmptyBody", ``, first(cltest.NewJobWithWebInitiator()), false},
		{"InvalidBody", `{`, first(cltest.NewJobWithWebInitiator()), true},
		{"WithoutWebInitiator", ``, first(cltest.NewJobWithLogInitiator()), true},
		{"NotFound", ``, first(cltest.NewJobWithWebInitiator()), true},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			assert.Nil(t, app.Store.SaveJob(&test.jobSpec))

			args := make([]string, 1)
			args[0] = test.jobSpec.ID
			if test.name == "NotFound" {
				args[0] = "badID"
			}

			if len(test.json) > 0 {
				args = append(args, test.json)
			}

			set := flag.NewFlagSet("run", 0)
			set.Parse(args)
			c := cli.NewContext(nil, set, nil)
			if test.errored {
				assert.Error(t, client.CreateJobRun(c))
			} else {
				assert.Nil(t, client.CreateJobRun(c))
			}
		})
	}
}

func TestClient_AddBridge(t *testing.T) {
	t.Parallel()
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	client, _ := app.NewClientAndRenderer()

	tests := []struct {
		name    string
		param   string
		errored bool
	}{
		{"EmptyString", "", true},
		{"ValidString", `{ "name": "TestBridge", "url": "http://localhost:3000/randomNumber" }`, false},
		{"InvalidString", `{ "noname": "", "nourl": "" }`, true},
		{"InvalidChar", `{ "badname": "path/bridge", "nourl": "" }`, true},
		{"ValidPath", "../internal/fixtures/web/create_random_number_bridge_type.json", false},
		{"InvalidPath", "bad/filepath/", true},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {

			set := flag.NewFlagSet("bridge", 0)
			set.Parse([]string{test.param})
			c := cli.NewContext(nil, set, nil)
			if test.errored {
				assert.Error(t, client.AddBridge(c))
			} else {
				assert.Nil(t, client.AddBridge(c))
			}
		})
	}
}

func TestClient_GetBridges(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	bt1 := &models.BridgeType{
		Name:          models.MustNewTaskType("testingbridges1"),
		URL:           cltest.WebURL("https://testing.com/bridges"),
		Confirmations: 0,
	}
	app.AddAdapter(bt1)

	bt2 := &models.BridgeType{
		Name:          models.MustNewTaskType("testingbridges2"),
		URL:           cltest.WebURL("https://testing.com/bridges"),
		Confirmations: 0,
	}
	app.AddAdapter(bt2)

	client, r := app.NewClientAndRenderer()

	require.Nil(t, client.GetBridges(cltest.EmptyCLIContext()))
	bridges := *r.Renders[0].(*[]models.BridgeType)
	require.Equal(t, 2, len(bridges))
	assert.Equal(t, bt1.Name, bridges[0].Name)
}

func TestClient_ShowBridge(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	bt := &models.BridgeType{
		Name:          models.MustNewTaskType("testingbridges1"),
		URL:           cltest.WebURL("https://testing.com/bridges"),
		Confirmations: 0,
	}
	app.AddAdapter(bt)

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{bt.Name.String()})
	c := cli.NewContext(nil, set, nil)
	require.Nil(t, client.ShowBridge(c))
	require.Equal(t, 1, len(r.Renders))
	assert.Equal(t, bt.Name, r.Renders[0].(*models.BridgeType).Name)
}

func TestClient_RemoveBridge(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	bt := &models.BridgeType{
		Name:          models.MustNewTaskType("testingbridges1"),
		URL:           cltest.WebURL("https://testing.com/bridges"),
		Confirmations: 0,
	}
	app.AddAdapter(bt)

	client, r := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{bt.Name.String()})
	c := cli.NewContext(nil, set, nil)
	require.Nil(t, client.RemoveBridge(c))
	require.Equal(t, 1, len(r.Renders))
	assert.Equal(t, bt.Name, r.Renders[0].(*models.BridgeType).Name)
}

func TestClient_RemoteLogin(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplication()
	defer cleanup()
	app.Start()

	tests := []struct {
		name, file string
		email, pwd string
		wantError  bool
	}{
		{"success prompt", "", cltest.APIEmail, cltest.Password, false},
		{"success file", "../internal/fixtures/apicredentials", "", "", false},
		{"failure prompt", "", "wrong@email.com", "wrongpwd", true},
		{"failure file", "/tmp/doesntexist", "", "", true},
		{"failure file w correct prompt", "/tmp/doesntexist", cltest.APIEmail, cltest.Password, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enteredStrings := []string{test.email, test.pwd}
			prompter := &cltest.MockCountingPrompter{EnteredStrings: enteredStrings}
			client := app.NewAuthenticatingClient(prompter)

			set := flag.NewFlagSet("test", 0)
			set.String("file", test.file, "")
			c := cli.NewContext(nil, set, nil)

			err := client.RemoteLogin(c)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_WithdrawSuccess(t *testing.T) {
	app, cleanup, _ := setupWithdrawalsApplication()
	defer cleanup()

	assert.NoError(t, app.StartAndConnect())

	client, _ := app.NewClientAndRenderer()
	set := flag.NewFlagSet("withdraw", 0)
	set.Parse([]string{"0x342156c8d3bA54Abc67920d35ba1d1e67201aC9C", "1"})

	c := cli.NewContext(nil, set, nil)

	assert.Nil(t, client.Withdraw(c))
}

func TestClient_WithdrawNoArgs(t *testing.T) {
	app, cleanup, _ := setupWithdrawalsApplication()
	defer cleanup()

	assert.NoError(t, app.StartAndConnect())

	client, _ := app.NewClientAndRenderer()
	set := flag.NewFlagSet("withdraw", 0)
	set.Parse([]string{})

	c := cli.NewContext(nil, set, nil)

	wr := client.Withdraw(c)
	assert.Error(t, wr)
	assert.Equal(t,
		"withdraw expects two arguments: an address and an amount",
		wr.Error())
}

func TestClient_WithdrawFromSpecifiedContractAddress(t *testing.T) {
	app, cleanup, ethMockCheck := setupWithdrawalsApplication()
	defer cleanup()

	assert.NoError(t, app.StartAndConnect())

	client, _ := app.NewClientAndRenderer()
	cliParserRouter := cmd.NewApp(client)
	assert.Nil(t, cliParserRouter.Run([]string{
		"chainlink", "withdraw",
		"0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF", "1234",
		"--from-oracle-contract-address=" +
			"0x3141592653589793238462643383279502884197"}))
	ethMockCheck(t)
}

func setupWithdrawalsApplication() (*cltest.TestApplication, func(), func(*testing.T)) {
	config, _ := cltest.NewConfig()
	oca := common.HexToAddress("0xDEADB3333333F")
	config.Set("ORACLE_CONTRACT_ADDRESS", &oca)
	app, cleanup := cltest.NewApplicationWithConfigAndKeyStore(config)

	hash := cltest.NewHash()
	sentAt := "0x5BA0"
	nonce := "0x100"
	ethMock := app.MockEthClient()

	ethMock.Context("app.Start()", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_getTransactionCount", nonce)
	})

	ethMock.Context("manager.CreateTx#1", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_call", "0xDE0B6B3A7640000")
		ethMock.Register("eth_sendRawTransaction", hash)
		ethMock.Register("eth_blockNumber", sentAt)
	})

	return app, cleanup, ethMock.EventuallyAllCalled
}

func first(a models.JobSpec, b interface{}) models.JobSpec {
	return a
}

func TestClient_SendEther(t *testing.T) {
	app, cleanup, _ := setupWithdrawalsApplication()
	defer cleanup()

	assert.NoError(t, app.StartAndConnect())

	client, _ := app.NewClientAndRenderer()
	set := flag.NewFlagSet("sendether", 0)
	set.Parse([]string{"100", "0x342156c8d3bA54Abc67920d35ba1d1e67201aC9C"})

	c := cli.NewContext(nil, set, nil)

	assert.Nil(t, client.SendEther(c))
}

func TestClient_ChangePassword(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	app.Start()

	enteredStrings := []string{cltest.APIEmail, cltest.Password}
	prompter := &cltest.MockCountingPrompter{EnteredStrings: enteredStrings}

	client := app.NewAuthenticatingClient(prompter)
	otherClient := app.NewAuthenticatingClient(prompter)

	set := flag.NewFlagSet("test", 0)
	set.String("file", "../internal/fixtures/apicredentials", "")
	c := cli.NewContext(nil, set, nil)
	err := client.RemoteLogin(c)
	assert.NoError(t, err)

	err = otherClient.RemoteLogin(c)
	assert.NoError(t, err)

	client.ChangePasswordPrompter = cltest.MockChangePasswordPrompter{
		ChangePasswordRequest: models.ChangePasswordRequest{
			OldPassword: cltest.Password,
			NewPassword: "password",
		},
	}
	err = client.ChangePassword(cli.NewContext(nil, nil, nil))
	assert.NoError(t, err)

	// otherClient should now be logged out
	err = otherClient.GetBridges(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401 Unauthorized")
}

func TestClient_GetTxAttempts(t *testing.T) {
	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	store := app.GetStore()
	from := cltest.GetAccountAddress(store)
	tx := cltest.CreateTxAndAttempt(store, from, 1)
	attempts, err := store.TxAttemptsFor(tx.ID)
	require.NoError(t, err)

	client, r := app.NewClientAndRenderer()

	// page 1
	set := flag.NewFlagSet("test txattempts", 0)
	set.Int("page", 1, "doc")
	c := cli.NewContext(nil, set, nil)
	require.Equal(t, 1, c.Int("page"))
	assert.NoError(t, client.GetTxAttempts(c))

	renderedAttempts := *r.Renders[0].(*[]models.TxAttempt)
	assert.Equal(t, 1, len(renderedAttempts))
	assert.Equal(t, attempts[0].Hash.Hex(), renderedAttempts[0].Hash.Hex())

	// page 2 which doesn't exist
	set = flag.NewFlagSet("test txattempts", 0)
	set.Int("page", 2, "doc")
	c = cli.NewContext(nil, set, nil)
	require.Equal(t, 2, c.Int("page"))
	assert.NoError(t, client.GetTxAttempts(c))

	renderedAttempts = *r.Renders[1].(*[]models.TxAttempt)
	assert.Equal(t, 0, len(renderedAttempts))
}

func TestClient_CreateExtraKey(t *testing.T) {
	app, cleanup := cltest.NewApplication()
	defer cleanup()
	app.Start()

	client, _ := app.NewClientAndRenderer()

	set := flag.NewFlagSet("test", 0)
	set.String("file", "../internal/fixtures/apicredentials", "")
	c := cli.NewContext(nil, set, nil)
	err := client.RemoteLogin(c)
	assert.NoError(t, err)

	client.PasswordPrompter = cltest.MockPasswordPrompter{Password: "password"}

	assert.NoError(t, client.CreateExtraKey(c))
}
