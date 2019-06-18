// Copyright 2017 The Cockroach Authors.
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

package distsqlrun

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/distsqlpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

// columnBackfiller is a processor for backfilling columns.
type columnBackfiller struct {
	backfiller

	backfill.ColumnBackfiller

	desc        *sqlbase.ImmutableTableDescriptor
	otherTables []*sqlbase.ImmutableTableDescriptor
}

var _ Processor = &columnBackfiller{}
var _ chunkBackfiller = &columnBackfiller{}

func newColumnBackfiller(
	flowCtx *FlowCtx,
	processorID int32,
	spec distsqlpb.BackfillerSpec,
	post *distsqlpb.PostProcessSpec,
	output RowReceiver,
) (*columnBackfiller, error) {
	otherTables := make([]*sqlbase.ImmutableTableDescriptor, len(spec.OtherTables))
	for i, tbl := range spec.OtherTables {
		otherTables[i] = sqlbase.NewImmutableTableDescriptor(tbl)
	}
	cb := &columnBackfiller{
		desc:        sqlbase.NewImmutableTableDescriptor(spec.Table),
		otherTables: otherTables,
		backfiller: backfiller{
			name:        "Column",
			filter:      backfill.ColumnMutationFilter,
			flowCtx:     flowCtx,
			processorID: processorID,
			output:      output,
			spec:        spec,
		},
	}
	cb.backfiller.chunks = cb

	if err := cb.ColumnBackfiller.Init(cb.flowCtx.NewEvalCtx(), cb.desc); err != nil {
		return nil, err
	}

	return cb, nil
}

func (cb *columnBackfiller) close(ctx context.Context) {}
func (cb *columnBackfiller) prepare(ctx context.Context) error {
	return nil
}
func (cb *columnBackfiller) flush(ctx context.Context) error {
	return nil
}
func (cb *columnBackfiller) CurrentBufferFill() float32 {
	return 0
}

// runChunk implements the chunkBackfiller interface.
func (cb *columnBackfiller) runChunk(
	ctx context.Context,
	mutations []sqlbase.DescriptorMutation,
	sp roachpb.Span,
	chunkSize int64,
	readAsOf hlc.Timestamp,
) (roachpb.Key, error) {
	var key roachpb.Key
	err := cb.flowCtx.ClientDB.Txn(ctx, func(ctx context.Context, txn *client.Txn) error {
		if cb.flowCtx.testingKnobs.RunBeforeBackfillChunk != nil {
			if err := cb.flowCtx.testingKnobs.RunBeforeBackfillChunk(sp); err != nil {
				return err
			}
		}
		if cb.flowCtx.testingKnobs.RunAfterBackfillChunk != nil {
			defer cb.flowCtx.testingKnobs.RunAfterBackfillChunk()
		}

		// TODO(knz): do KV tracing in DistSQL processors.
		var err error
		key, err = cb.RunColumnBackfillChunk(
			ctx,
			txn,
			cb.desc,
			cb.otherTables,
			sp,
			chunkSize,
			true,  /*alsoCommit*/
			false, /*traceKV*/
		)
		return err
	})
	return key, err
}