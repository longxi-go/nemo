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

import (
	"fmt"
	"os"
)

const (
	nonRegularOutputFormat                          = "generated output destination is not a regular file: %s"
	refusingToReplaceUnownedGeneratedFilenameFormat = "refusing to replace unowned generated filename: %s"
)

// Discovery must enforce output ownership before overlays hide an existing file.
// All read-only modes use this check too; a successful check must not endorse a
// symlink destination that the shared committer would refuse to replace.
func readGeneratedOutput(path string) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf(nonRegularOutputFormat, path)
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	if !ownsGeneratedOutput(path, current) {
		return nil, false, fmt.Errorf(refusingToReplaceUnownedGeneratedFilenameFormat, path)
	}
	return current, true, nil
}
