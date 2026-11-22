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
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestYAMLPreflightSharedAliasesHaveBoundedWork(t *testing.T) {
	leaf := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "7",
	}
	root := leaf
	for range 12 {
		root = &yaml.Node{
			Kind: yaml.SequenceNode,
			Content: []*yaml.Node{
				{
					Kind:  yaml.AliasNode,
					Alias: root,
				},
				{
					Kind:  yaml.AliasNode,
					Alias: root,
				},
			},
		}
	}
	var validationErr error
	allocations := testing.AllocsPerRun(1, func() {
		validationErr = validateYAMLNumbers(root, 0)
	})
	if validationErr != nil {
		t.Fatal(validationErr)
	}
	// The graph has 37 nodes. Rechecking all expanded paths allocates thousands
	// of scalar parse objects even though aliases contain no new numeric text.
	if allocations > 512 {
		t.Fatalf("shared aliases were repeatedly expanded: %.0f allocations", allocations)
	}
}

func TestYAMLPreflightChecksAliasDepthAtEachUse(t *testing.T) {
	leaf := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "7",
	}
	deep := leaf
	for range maxDocumentDepth {
		deep = &yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: []*yaml.Node{deep},
		}
	}
	root := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: []*yaml.Node{leaf, deep},
	}
	if err := validateYAMLNumbers(root, 0); err == nil {
		t.Fatal("a prior shallow visit bypassed the deep occurrence limit")
	}
}

func TestYAMLPreflightRejectsAliasCycles(t *testing.T) {
	root := &yaml.Node{
		Kind: yaml.SequenceNode,
	}
	root.Content = []*yaml.Node{
		{
			Kind:  yaml.AliasNode,
			Alias: root,
		},
	}
	if err := validateYAMLNumbers(root, 0); err == nil {
		t.Fatal("alias cycle accepted")
	}
}
