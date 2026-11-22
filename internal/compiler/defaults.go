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
	"encoding/csv"
	"fmt"
	"go/types"
	"math"
	"strconv"
	"strings"
	"time"
)

func validateDefault(t types.Type, raw string, sizes types.Sizes) error {
	t = types.Unalias(t)
	if named, ok := t.(*types.Named); ok {
		if hasText(named) {
			return nil
		}
		if isDurationType(named) {
			_, err := time.ParseDuration(raw)
			return err
		}
	}
	switch typ := t.Underlying().(type) {
	case *types.Pointer:
		return validateDefault(typ.Elem(), raw, sizes)
	case *types.Slice:
		if raw == "" {
			return nil
		}
		records, err := csv.NewReader(strings.NewReader(raw)).ReadAll()
		if err != nil || len(records) != 1 {
			return fmt.Errorf("list default requires one CSV record")
		}
		for _, item := range records[0] {
			if err := validateDefault(typ.Elem(), item, sizes); err != nil {
				return err
			}
		}
		return nil
	case *types.Basic:
		bits := 64
		switch typ.Kind() {
		case types.Int8, types.Uint8:
			bits = 8
		case types.Int16, types.Uint16:
			bits = 16
		case types.Int32, types.Uint32, types.Float32:
			bits = 32
		case types.Int, types.Uint, types.Uintptr:
			if sizes == nil {
				return fmt.Errorf("target integer width is unavailable")
			}
			bits = int(sizes.Sizeof(typ) * 8)
		}
		switch {
		case typ.Info()&types.IsString != 0:
			return nil
		case typ.Info()&types.IsBoolean != 0:
			_, err := strconv.ParseBool(raw)
			return err
		case typ.Info()&types.IsUnsigned != 0:
			_, err := strconv.ParseUint(raw, 10, bits)
			return err
		case typ.Info()&types.IsInteger != 0:
			_, err := strconv.ParseInt(raw, 10, bits)
			return err
		case typ.Info()&types.IsFloat != 0:
			value, err := strconv.ParseFloat(raw, bits)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("invalid finite float default")
			}
			return nil
		}
	}
	return fmt.Errorf("defaults require scalar or scalar-list fields")
}
