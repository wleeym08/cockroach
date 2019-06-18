// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package container

import (
	"context"
	"errors"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/closedts"
	"github.com/cockroachdb/cockroach/pkg/storage/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type noopEverything struct{}

// NoopContainer returns a Container for which all parts of the subsystem are
// mocked out. This is for usage in testing where there just needs to be a
// structure that stays out of the way.
//
// The returned container will behave correctly. It will never allow any time-
// stamps to be closed out, so it never makes any promises; it doesn't use any
// locking and it does not consume any (nontrivial) resources.
func NoopContainer() *Container {
	return &Container{
		Config: Config{
			Settings: cluster.MakeTestingClusterSettings(),
			Stopper:  stop.NewStopper(),
			Clock: func(roachpb.NodeID) (hlc.Timestamp, ctpb.Epoch, error) {
				return hlc.Timestamp{}, 0, errors.New("closed timestamps disabled for testing")
			},
			Refresh: func(...roachpb.RangeID) {},
			Dialer:  noopEverything{},
		},
		Tracker:  noopEverything{},
		Storage:  noopEverything{},
		Provider: noopEverything{},
		Server:   noopEverything{},
		Clients:  noopEverything{},
		noop:     true,
	}
}

func (noopEverything) Get(client ctpb.InboundClient) error {
	return errors.New("closed timestamps disabled")
}
func (noopEverything) Close(
	next hlc.Timestamp, expCurEpoch ctpb.Epoch,
) (hlc.Timestamp, map[roachpb.RangeID]ctpb.LAI, bool) {
	return hlc.Timestamp{}, nil, false
}
func (noopEverything) Track(ctx context.Context) (hlc.Timestamp, closedts.ReleaseFunc) {
	return hlc.Timestamp{}, func(context.Context, ctpb.Epoch, roachpb.RangeID, ctpb.LAI) {}
}
func (noopEverything) VisitAscending(roachpb.NodeID, func(ctpb.Entry) (done bool))  {}
func (noopEverything) VisitDescending(roachpb.NodeID, func(ctpb.Entry) (done bool)) {}
func (noopEverything) Add(roachpb.NodeID, ctpb.Entry)                               {}
func (noopEverything) Clear()                                                       {}
func (noopEverything) Notify(roachpb.NodeID) chan<- ctpb.Entry {
	return nil // will explode when used, but nobody would use this
}
func (noopEverything) Subscribe(context.Context, chan<- ctpb.Entry) {}
func (noopEverything) Start()                                       {}
func (noopEverything) MaxClosed(
	roachpb.NodeID, roachpb.RangeID, ctpb.Epoch, ctpb.LAI,
) hlc.Timestamp {
	return hlc.Timestamp{}
}
func (noopEverything) Request(roachpb.NodeID, roachpb.RangeID) {}
func (noopEverything) EnsureClient(roachpb.NodeID)             {}
func (noopEverything) Dial(context.Context, roachpb.NodeID) (ctpb.Client, error) {
	return nil, errors.New("closed timestamps disabled")
}
func (noopEverything) Ready(roachpb.NodeID) bool { return false }