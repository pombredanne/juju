// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package backups

import (
	"io"

	"github.com/juju/errors"
	"github.com/juju/loggo"

	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/state"
	"github.com/juju/juju/state/backups"
)

func init() {
	common.RegisterStandardFacade("Backups", 0, NewAPI)
}

// TODO(ericsnow) lp-1389362
// The machine ID needs to be introspected from the API server, likely
// through a Resource.
const machineID = "0"

var logger = loggo.GetLogger("juju.apiserver.backups")

// API serves backup-specific API methods.
type API struct {
	st        *state.State
	paths     *backups.Paths
	machineID string
}

// NewAPI creates a new instance of the Backups API facade.
func NewAPI(st *state.State, resources *common.Resources, authorizer common.Authorizer) (*API, error) {
	if !authorizer.AuthClient() {
		return nil, errors.Trace(common.ErrPerm)
	}

	// Get the backup paths.

	dataDirRes := resources.Get("dataDir")
	dataDir, ok := dataDirRes.(common.StringResource)
	if !ok {
		if dataDirRes == nil {
			dataDir = ""
		} else {
			return nil, errors.Errorf("invalid dataDir resource: %v", dataDirRes)
		}
	}

	logDirRes := resources.Get("logDir")
	logDir, ok := logDirRes.(common.StringResource)
	if !ok {
		if logDirRes == nil {
			logDir = ""
		} else {
			return nil, errors.Errorf("invalid logDir resource: %v", logDirRes)
		}
	}

	paths := backups.Paths{
		DataDir: dataDir.String(),
		LogsDir: logDir.String(),
	}

	// Build the API.
	b := API{
		st:        st,
		paths:     &paths,
		machineID: machineID,
	}
	return &b, nil
}

var newBackups = func(st *state.State) (backups.Backups, io.Closer) {
	stor := state.NewBackupStorage(st)
	return backups.NewBackups(stor), stor
}
