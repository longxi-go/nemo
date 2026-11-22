// Copyright © 2026-present the longxi-go/nemo authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package admission

import (
	"context"
	"math"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/limits"
)

// Budget counts resource use for one candidate. It checks values before
// adding traversal work so wide collections cannot exceed the node limit first.
type Budget struct {
	ctx          context.Context
	policy       limits.Policy
	nodes        int64
	bytes        int64
	byteResource limits.Resource
	byteLimit    int64
}

func New(ctx context.Context, policy limits.Policy, resolved bool) *Budget {
	resource, maximum := limits.ValueBytes, policy.MaxValueBytes
	if resolved {
		resource, maximum = limits.ResolvedBytes, policy.MaxResolvedBytes
	}
	return &Budget{
		ctx:          ctx,
		policy:       policy,
		byteResource: resource,
		byteLimit:    maximum,
	}
}

func (b *Budget) Bytes(count int) error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if int64(count) > b.byteLimit-b.bytes {
		observed := int64(math.MaxInt64)
		if int64(count) <= math.MaxInt64-b.bytes {
			observed = b.bytes + int64(count)
		}
		return &limits.ExceededError{
			Resource: b.byteResource,
			Limit:    b.byteLimit,
			Observed: observed,
		}
	}
	b.bytes += int64(count)
	return nil
}

func (b *Budget) Property(key string, value gosr.Value) error {
	if err := b.Bytes(len(key)); err != nil {
		return err
	}
	return b.Value(value)
}

type frame struct {
	value gosr.Value
	depth int
}

func (b *Budget) admit(value gosr.Value, depth int) error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if err := limits.Check(limits.ValueDepth, int64(depth), int64(b.policy.MaxValueDepth)); err != nil {
		return err
	}
	if err := limits.Check(limits.ValueNodes, b.nodes+1, int64(b.policy.MaxValueNodes)); err != nil {
		return err
	}
	b.nodes++
	if text, ok := value.Scalar(); ok {
		return b.Bytes(len(text))
	}
	return nil
}

func (b *Budget) Value(value gosr.Value) error {
	if err := b.admit(value, 0); err != nil {
		return err
	}
	stack := []frame{
		{
			value: value,
		},
	}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if err := limits.Check(limits.ValueNodes, b.nodes+int64(current.value.Len()), int64(b.policy.MaxValueNodes)); err != nil {
			return err
		}
		for _, child := range current.value.Elements() {
			if err := b.admit(child, current.depth+1); err != nil {
				return err
			}
			stack = append(stack, frame{
				value: child,
				depth: current.depth + 1,
			})
		}
		for name, child := range current.value.Members() {
			if err := b.Bytes(len(name)); err != nil {
				return err
			}
			if err := b.admit(child, current.depth+1); err != nil {
				return err
			}
			stack = append(stack, frame{
				value: child,
				depth: current.depth + 1,
			})
		}
	}
	return nil
}
