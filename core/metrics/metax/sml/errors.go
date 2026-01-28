// Copyright 2025 The HuaTuo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sml

import (
	"errors"
	"fmt"
)

type SmlError struct {
	fn   string
	Code Return
}

func (e *SmlError) Error() string {
	return fmt.Sprintf("%s failed: %s", e.fn, e.Code.String())
}

func IsNotSupported(err error) bool {
	var smlErr *SmlError
	return errors.As(err, &smlErr) &&
		smlErr.Code == ERROR_OPERATION_NOT_SUPPORTED
}

func checkReturnCode(operation string, code Return) error {
	if code == SUCCESS {
		return nil
	}

	return &SmlError{
		fn:   operation,
		Code: code,
	}
}
