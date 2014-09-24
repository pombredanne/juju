// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package firewaller

import (
	"github.com/juju/errors"
	"github.com/juju/names"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/api/common"
	"github.com/juju/juju/api/watcher"
	"github.com/juju/juju/apiserver/params"
)

const firewallerFacade = "Firewaller"

// State provides access to the Firewaller API facade.
type State struct {
	facade base.FacadeCaller
	*common.EnvironWatcher
}

// newStateForVersion creates a new client-side Firewaller API facade
// for the given version. If version is -1, the best facade version
// among the ones supported by both the client and the server is
// chosen.
func newStateForVersion(caller base.APICaller, version int) *State {
	var facadeCaller base.FacadeCaller
	if version == -1 {
		facadeCaller = base.NewFacadeCaller(caller, firewallerFacade)
	} else {
		facadeCaller = base.NewFacadeCallerForVersion(caller, firewallerFacade, version)
	}
	return &State{
		facade:         facadeCaller,
		EnvironWatcher: common.NewEnvironWatcher(facadeCaller),
	}
}

// newStateV0 creates a new client-side Firewaller facade, version 0.
func newStateV0(caller base.APICaller) *State {
	return newStateForVersion(caller, 0)
}

// newStateV1 creates a new client-side Firewaller facade, version 1.
func newStateV1(caller base.APICaller) *State {
	return newStateForVersion(caller, 1)
}

// newStateBestVersion creates a new client-side Firewaller facade
// with the best API version supported by both the client and the
// server.
func newStateBestVersion(caller base.APICaller) *State {
	return newStateForVersion(caller, -1)
}

// NewState creates a new client-side Firewaller facade.
// Defined like this to allow patching during tests.
var NewState = newStateBestVersion

// BestAPIVersion returns the API version that we were able to
// determine is supported by both the client and the API Server.
func (st *State) BestAPIVersion() int {
	return st.facade.BestAPIVersion()
}

// EnvironTag returns the current environment's tag.
func (st *State) EnvironTag() (names.EnvironTag, error) {
	return st.facade.RawAPICaller().EnvironTag()
}

// life requests the life cycle of the given entity from the server.
func (st *State) life(tag names.Tag) (params.Life, error) {
	return common.Life(st.facade, tag)
}

// Unit provides access to methods of a state.Unit through the facade.
func (st *State) Unit(tag names.UnitTag) (*Unit, error) {
	life, err := st.life(tag)
	if err != nil {
		return nil, err
	}
	return &Unit{
		tag:  tag,
		life: life,
		st:   st,
	}, nil
}

// Machine provides access to methods of a state.Machine through the
// facade.
func (st *State) Machine(tag names.MachineTag) (*Machine, error) {
	life, err := st.life(tag)
	if err != nil {
		return nil, err
	}
	return &Machine{
		tag:  tag,
		life: life,
		st:   st,
	}, nil
}

// WatchEnvironMachines returns a StringsWatcher that notifies of
// changes to the life cycles of the top level machines in the current
// environment.
func (st *State) WatchEnvironMachines() (watcher.StringsWatcher, error) {
	var result params.StringsWatchResult
	err := st.facade.FacadeCall("WatchEnvironMachines", nil, &result)
	if err != nil {
		return nil, err
	}
	if err := result.Error; err != nil {
		return nil, result.Error
	}
	w := watcher.NewStringsWatcher(st.facade.RawAPICaller(), result)
	return w, nil
}

// WatchOpenedPorts returns a StringsWatcher that notifies of
// changes to the opened ports for the current environment.
func (st *State) WatchOpenedPorts() (watcher.StringsWatcher, error) {
	if st.BestAPIVersion() < 1 {
		// WatchOpenedPorts() was introduced in FirewallerAPIV1.
		return nil, errors.NotImplementedf("WatchOpenedPorts() (need V1+)")
	}
	envTag, err := st.EnvironTag()
	if err != nil {
		return nil, errors.Annotatef(err, "invalid environ tag")
	}
	var results params.StringsWatchResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: envTag.String()}},
	}
	err = st.facade.FacadeCall("WatchOpenedPorts", args, &results)
	if err != nil {
		return nil, err
	}
	if len(results.Results) != 1 {
		return nil, errors.Errorf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if err := result.Error; err != nil {
		return nil, result.Error
	}
	w := watcher.NewStringsWatcher(st.facade.RawAPICaller(), result)
	return w, nil
}