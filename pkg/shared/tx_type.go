// VulcanizeDB
// Copyright © 2020 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package shared

import (
	"fmt"
	"strings"
)

type TxType int

const (
	Unsupported TxType = iota
	Standard
	OptimismL2
	OptimismL1ToL2
	EIP1559
)

func (tt TxType) String() string {
	switch tt {
	case Standard:
		return "Standard"
	case OptimismL2:
		return "L2"
	case OptimismL1ToL2:
		return "L1toL2"
	case EIP1559:
		return "EIP1559"
	default:
		return "Unsupported"
	}
}

// TxTypeFromString returns the tx enum type from provided string
func TxTypeFromString(str string) (TxType, error) {
	switch strings.ToLower(str) {
	case "geth", "standard", "parity":
		return Standard, nil
	case "l2":
		return OptimismL2, nil
	case "l1", "l2tol1":
		return OptimismL1ToL2, nil
	case "eip1559":
		return EIP1559, nil
	default:
		return Unsupported, fmt.Errorf("unsupported tx type: %s", str)
	}
}
