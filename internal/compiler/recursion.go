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

package compiler

import "github.com/longxi-go/nemo/pkg/maptypes"

// markRecursivePointers tests each edge against the complete graph, not only a
// DFS back edge. This keeps mutual recursion and multiple roots independent of
// traversal/declaration order. Collection edges participate in reachability but
// need no absence gate: missing/empty collections already terminate traversal.
func markRecursivePointers(models []helperModel) {
	edges := make(maptypes.StringMap[[]string], len(models))
	for _, model := range models {
		for _, field := range model.Fields {
			target := objectEdge(field)
			if target != "" {
				edges[model.Name] = append(edges[model.Name], target)
			}
		}
	}
	for i := range models {
		for j := range models[i].Fields {
			field := &models[i].Fields[j]
			field.Recursive = field.Mode == fieldModeLocal && field.Pointer && reachable(edges, field.Helper, models[i].Name)
		}
	}
}

func objectEdge(field fieldModel) string {
	if field.Mode == fieldModeLocal {
		return field.Helper
	}
	if field.Element != nil && !field.Element.External {
		return field.Element.Helper
	}
	return ""
}

func reachable(edges maptypes.StringMap[[]string], start, goal string) bool {
	pending := []string{start}
	seen := make(maptypes.StringBoolMap)
	for len(pending) > 0 {
		node := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if node == goal {
			return true
		}
		if seen[node] {
			continue
		}
		seen[node] = true
		pending = append(pending, edges[node]...)
	}
	return false
}
