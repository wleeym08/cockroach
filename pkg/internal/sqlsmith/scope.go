// Copyright 2019 The Cockroach Authors.
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

package sqlsmith

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

// colRef refers to a named result column. If it is from a table, def is
// populated.
type colRef struct {
	typ  *types.T
	item *tree.ColumnItem
}

type colRefs []*colRef

func (t colRefs) extend(refs ...*colRef) colRefs {
	ret := append(make(colRefs, 0, len(t)+len(refs)), t...)
	ret = append(ret, refs...)
	return ret
}

func (t colRefs) stripTableName() {
	for _, c := range t {
		c.item.TableName = nil
	}
}

type scope struct {
	schema *Smither

	// The budget tracks available complexity. It is randomly generated. Each
	// call to canRecurse decreases it such that canRecurse will eventually
	// always return false.
	budget int
}

func (s *Smither) makeScope() *scope {
	return &scope{
		schema: s,
		budget: s.rnd.Intn(100),
	}
}

// canRecurse returns whether the current function should possibly invoke
// a function that calls creates new nodes.
func (s *scope) canRecurse() bool {
	s.budget--
	// Disable recursion randomly so that early expressions don't take all
	// the budget.
	return s.budget > 0 && s.coin()
}

// Context holds information about what kinds of expressions are legal at
// a particular place in a query.
type Context struct {
	fnClass  tree.FunctionClass
	noWindow bool
}

var (
	emptyCtx   = Context{}
	groupByCtx = Context{fnClass: tree.AggregateClass}
	havingCtx  = Context{
		fnClass:  tree.AggregateClass,
		noWindow: true,
	}
	windowCtx = Context{fnClass: tree.WindowClass}
)