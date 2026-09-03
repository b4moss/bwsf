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
	exitFunc         = os.Exit
	confirmOverwrite = utils.ConfirmOverwrite
	selectCleanMismatch = utils.SelectCleanMismatchAction
	inputPassword    = utils.InputPassword
)
