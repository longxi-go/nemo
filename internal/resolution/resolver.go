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

package resolution

import (
	"context"
	"errors"
	"maps"
	"slices"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	maxDepth  = 64
	maxOutput = 1 << 20
)

// Resolver stores a fully resolved property map. Construction copies the input
// map; subsequent reads perform no I/O and do not expand escaped text again.
type Resolver struct {
	properties maptypes.PropertyMap
	heights    maptypes.StringMap[int]
	keys       []string
}

// New resolves all selected properties using the default resource limits.
func New(raw maptypes.PropertyMap) (*Resolver, error) {
	return NewWithPolicy(context.Background(), raw, limits.Default())
}

// NewWithPolicy applies separate limits to raw values and expanded output.
func NewWithPolicy(ctx context.Context, raw maptypes.PropertyMap, policy limits.Policy) (*Resolver, error) {
	if ctx == nil {
		return nil, failure(gosr.InvalidState)
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if err := limits.Check(limits.Properties, int64(len(raw)), int64(policy.MaxProperties)); err != nil {
		return nil, err
	}
	input := admission.New(ctx, policy, false)
	for key, property := range raw {
		if err := input.Property(key, property.Value); err != nil {
			return nil, err
		}
	}
	r := &Resolver{
		properties: make(maptypes.PropertyMap, len(raw)),
		heights:    make(maptypes.StringMap[int], len(raw)),
	}
	e := engine{
		ctx:    ctx,
		output: admission.New(ctx, policy, true),
		raw:    maps.Clone(raw),
		result: r,
		active: maptypes.StringBoolMap{},
	}
	for key := range raw {
		if err := gosr.ValidateKey(key); err != nil {
			return nil, err
		}
		r.keys = append(r.keys, key)
	}
	slices.Sort(r.keys)
	for _, key := range r.keys {
		if _, _, err := e.property(key); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Lookup returns resolved immutable values and retained provenance without
// parsing substitutions again. Nil resolvers and invalid keys return classified errors.
func (r *Resolver) Lookup(key string) (gosr.Property, bool, error) {
	if r == nil {
		return gosr.Property{}, false, failure(gosr.InvalidState)
	}
	if err := gosr.ValidateKey(key); err != nil {
		return gosr.Property{}, false, err
	}
	p, ok := r.properties[key]
	return p, ok, nil
}

// Keys detaches sorted addresses; a nil resolver returns no keys.
func (r *Resolver) Keys() []string {
	if r == nil {
		return nil
	}
	return slices.Clone(r.keys)
}

// Resolve expands caller text against the pinned generation, retaining the
// original expansion heights so caching cannot bypass nesting limits.
func (r *Resolver) Resolve(text string) (string, error) {
	if r == nil {
		return "", failure(gosr.InvalidState)
	}
	p := parser{
		operand: r.operand,
	}
	out, _, err := p.expand(text)
	return out, err
}

func (r *Resolver) operand(key string) (gosr.Value, int, bool, error) {
	v, ok := r.properties[key]
	return v.Value, r.heights[key], ok, nil
}

type engine struct {
	ctx    context.Context
	output *admission.Budget
	raw    maptypes.PropertyMap
	result *Resolver
	active maptypes.StringBoolMap
	path   []string
}

func (e *engine) property(key string) (gosr.Value, int, error) {
	if err := e.ctx.Err(); err != nil {
		return gosr.Value{}, 0, err
	}
	if e.active[key] {
		return gosr.Value{}, 0, &gosr.Error{
			Code: gosr.ResolutionCycle,
			Key:  key,
			Path: append(slices.Clone(e.path), key),
		}
	}
	if p, ok := e.result.properties[key]; ok {
		return p.Value, e.result.heights[key], nil
	}
	if len(e.path) > maxDepth {
		return gosr.Value{}, 0, &gosr.Error{
			Code: gosr.ResolutionLimit,
			Key:  key,
		}
	}
	p := e.raw[key]
	e.active[key] = true
	e.path = append(e.path, key)
	err := e.output.Bytes(len(key))
	var v gosr.Value
	var h int
	if err == nil {
		v, h, err = e.value(p.Value)
	}
	path := slices.Clone(e.path)
	delete(e.active, key)
	e.path = e.path[:len(e.path)-1]
	if err != nil {
		var detail *gosr.Error
		if errors.As(err, &detail) {
			if len(detail.Path) > 0 {
				path = slices.Clone(detail.Path)
			} else if detail.Key != "" {
				path = append(path, detail.Key)
			}
		}
		return gosr.Value{}, 0, &gosr.Error{
			Code:   codeOf(err),
			Key:    key,
			Origin: p.Origin,
			Path:   path,
			Cause:  err,
		}
	}
	p.Value = v
	e.result.properties[key] = p
	e.result.heights[key] = h
	return v, h, nil
}

func (e *engine) value(v gosr.Value) (gosr.Value, int, error) {
	if err := e.ctx.Err(); err != nil {
		return gosr.Value{}, 0, err
	}
	switch v.Kind() {
	case gosr.StringKind:
		text, _ := v.Scalar()
		p := parser{
			operand: e.operand,
			check:   e.ctx.Err,
			charge:  e.output.Bytes,
		}
		out, h, err := p.expand(text)
		return gosr.String(out), h, err
	case gosr.ListKind:
		values := v.Items()
		height := 0
		for i, item := range values {
			next, h, err := e.value(item)
			if err != nil {
				return gosr.Value{}, 0, err
			}
			values[i] = next
			height = max(height, h)
		}
		return gosr.List(values...), height, nil
	case gosr.ObjectKind:
		fields := v.Fields()
		height := 0
		for key, item := range fields {
			if err := e.output.Bytes(len(key)); err != nil {
				return gosr.Value{}, 0, err
			}
			next, h, err := e.value(item)
			if err != nil {
				return gosr.Value{}, 0, err
			}
			fields[key] = next
			height = max(height, h)
		}
		return gosr.Object(fields), height, nil
	default:
		if text, ok := v.Scalar(); ok {
			if err := e.output.Bytes(len(text)); err != nil {
				return gosr.Value{}, 0, err
			}
		}
		return v, 0, nil
	}
}

func (e *engine) operand(key string) (gosr.Value, int, bool, error) {
	if _, ok := e.raw[key]; !ok {
		return gosr.Value{}, 0, false, nil
	}
	v, h, err := e.property(key)
	return v, h, true, err
}
