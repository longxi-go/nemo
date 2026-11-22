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

package codec

import (
	"math/big"
	"strings"

	"go.yaml.in/yaml/v3"
)

const yamlDigitSeparator = "_"

type yamlVisit struct {
	node  *yaml.Node
	depth int
}

type yamlPreflight struct {
	visited map[yamlVisit]bool
	active  map[*yaml.Node]bool
}

// yaml.v3 may decode overflowing integer text as float64. Validate numeric text
// before decoding; memoize shared nodes at each depth rather than expanding the
// alias graph before the library can apply its own alias-ratio protection.
func validateYAMLNumbers(node *yaml.Node, depth int) error {
	validation := yamlPreflight{
		visited: make(map[yamlVisit]bool),
		active:  make(map[*yaml.Node]bool),
	}
	return validation.visit(node, depth)
}

func (v *yamlPreflight) visit(node *yaml.Node, depth int) error {
	if node == nil || depth > maxDocumentDepth || v.active[node] {
		return decodeError(nil)
	}
	key := yamlVisit{
		node:  node,
		depth: depth,
	}
	if v.visited[key] {
		return nil
	}
	v.active[node] = true
	defer delete(v.active, node)
	if node.Kind == yaml.AliasNode {
		if err := v.visit(node.Alias, depth); err != nil {
			return err
		}
	} else {
		if err := validateYAMLScalar(node); err != nil {
			return err
		}
		for index, child := range node.Content {
			if node.Kind == yaml.MappingNode && index%2 == 0 {
				continue
			}
			if err := v.visit(child, depth+1); err != nil {
				return err
			}
		}
	}
	v.visited[key] = true
	return nil
}

func validateYAMLScalar(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.Style != 0 {
		return nil
	}
	text := strings.ReplaceAll(node.Value, yamlDigitSeparator, "")
	integer, ok := new(big.Int).SetString(text, 0)
	if !ok {
		integer, ok = new(big.Int).SetString(text, 10)
	}
	if ok && !integer.IsInt64() && !integer.IsUint64() {
		return decodeError(nil)
	}
	return nil
}
