//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type VgOptions struct {
	SinglerowID          bool `sql:"primary_key"`
	LastReconciliationAt *time.Time
}
