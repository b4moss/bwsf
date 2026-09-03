package cmd

import "bwsf/src/config"

// Version is the current version of bwsf
const Version = "0.20.0"

func init() {
	config.AppVersion = Version
}
