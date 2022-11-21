// Code generated by "enumer -type=TransactionDirection -json"; DO NOT EDIT.

package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _TransactionDirectionName = "InboundOutbound"

var _TransactionDirectionIndex = [...]uint8{0, 7, 15}

const _TransactionDirectionLowerName = "inboundoutbound"

func (i TransactionDirection) String() string {
	i -= 1
	if i < 0 || i >= TransactionDirection(len(_TransactionDirectionIndex)-1) {
		return fmt.Sprintf("TransactionDirection(%d)", i+1)
	}
	return _TransactionDirectionName[_TransactionDirectionIndex[i]:_TransactionDirectionIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _TransactionDirectionNoOp() {
	var x [1]struct{}
	_ = x[Inbound-(1)]
	_ = x[Outbound-(2)]
}

var _TransactionDirectionValues = []TransactionDirection{Inbound, Outbound}

var _TransactionDirectionNameToValueMap = map[string]TransactionDirection{
	_TransactionDirectionName[0:7]:       Inbound,
	_TransactionDirectionLowerName[0:7]:  Inbound,
	_TransactionDirectionName[7:15]:      Outbound,
	_TransactionDirectionLowerName[7:15]: Outbound,
}

var _TransactionDirectionNames = []string{
	_TransactionDirectionName[0:7],
	_TransactionDirectionName[7:15],
}

// TransactionDirectionString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func TransactionDirectionString(s string) (TransactionDirection, error) {
	if val, ok := _TransactionDirectionNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _TransactionDirectionNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to TransactionDirection values", s)
}

// TransactionDirectionValues returns all values of the enum
func TransactionDirectionValues() []TransactionDirection {
	return _TransactionDirectionValues
}

// TransactionDirectionStrings returns a slice of all String values of the enum
func TransactionDirectionStrings() []string {
	strs := make([]string, len(_TransactionDirectionNames))
	copy(strs, _TransactionDirectionNames)
	return strs
}

// IsATransactionDirection returns "true" if the value is listed in the enum definition. "false" otherwise
func (i TransactionDirection) IsATransactionDirection() bool {
	for _, v := range _TransactionDirectionValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for TransactionDirection
func (i TransactionDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for TransactionDirection
func (i *TransactionDirection) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TransactionDirection should be a string, got %s", data)
	}

	var err error
	*i, err = TransactionDirectionString(s)
	return err
}
