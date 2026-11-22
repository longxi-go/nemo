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

package environment

import "context"

// refreshGate serializes candidate preparation and lets queued callers cancel.
// Each admitted Refresh releases the token before returning.
type refreshGate chan struct{}

func (g refreshGate) acquire(ctx context.Context) error {
	if ctx == nil {
		return invalid()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case g <- struct{}{}:
		if err := ctx.Err(); err != nil {
			g.release()
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g refreshGate) release() {
	<-g
}
