//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import "errors"

type Broadcastertype string

const (
	Broadcastertype_Partner   Broadcastertype = "partner"
	Broadcastertype_Affiliate Broadcastertype = "affiliate"
	Broadcastertype_Normal    Broadcastertype = "normal"
)

func (e *Broadcastertype) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("jet: Invalid scan value for AllTypesEnum enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "partner":
		*e = Broadcastertype_Partner
	case "affiliate":
		*e = Broadcastertype_Affiliate
	case "normal":
		*e = Broadcastertype_Normal
	default:
		return errors.New("jet: Invalid scan value '" + enumValue + "' for Broadcastertype enum")
	}

	return nil
}

func (e Broadcastertype) String() string {
	return string(e)
}