package cmd

import (
	"os"

	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
)

// Package-level deps for unit tests (swap + t.Cleanup). Production defaults call real infra/utils.
var (
	checkBwInstalled = utils.CheckBwCommand
	newBwClient      = func() core.BwClient { return infra.NewBwClient() }
	newFileSystem    = func() core.FileSystem { return infra.NewFileSystem() }
	newLogger        = func() core.Logger { return infra.NewLogger() }
	newSessionStore  = func() core.SessionStore { return infra.NewSessionStore() }
	newSecretStore   = func() infra.SecretStore { return infra.NewKeyringStore() }
	newUnlockClient  = func(cfg *config.Config, host *config.Host, store infra.SecretStore) unlockClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, store, infra.NewIdentityClient(), infra.NewSDKCryptoSession())
	}
	exitFunc            = os.Exit
	confirmOverwrite    = utils.ConfirmOverwrite
	selectCleanMismatch = utils.SelectCleanMismatchAction
	inputPassword       = utils.InputPassword
)

// unlockClient is the subset of ApiBwClient used by unlock/lock.
type unlockClient interface {
	Unlock(masterPassword string) error
	ClearSession()
}
