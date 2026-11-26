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

package binding

import (
	"strconv"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

// Index markers describe diagnostic paths, not addressable property keys.
const (
	listIndexPrefix = "["
	listIndexSuffix = "]"
)

// ListObjects accepts an exact list of objects. Missing optional input is nil;
// an explicit empty list is non-nil. CSV and flat numeric keys are not object lists.
func ListObjects(r gosr.PropertyResolver, key string, required bool) ([]ObjectInput, error) {
	if r == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	property, found, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	if !found {
		present, err := Present(r, key)
		if err != nil {
			return nil, err
		}
		if present {
			return nil, objectFailure(key, gosr.Origin{})
		}
		if required {
			return nil, &gosr.Error{
				Code: gosr.MissingProperty,
				Key:  key,
			}
		}
		return nil, nil
	}
	if property.Value.Kind() != gosr.ListKind {
		return nil, objectFailure(key, property.Origin)
	}
	if err := admitComposite(key, property); err != nil {
		return nil, err
	}
	items := property.Value.Items()
	inputs := make([]ObjectInput, 0, len(items))
	for i, value := range items {
		indexPath := listIndexPrefix + strconv.Itoa(i) + listIndexSuffix
		input, err := newObjectInput(r, objectElement{
			prefix:    key,
			value:     value,
			sourceKey: key,
			origin:    property.Origin,
			path:      indexPath,
		})
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	return inputs, nil
}
