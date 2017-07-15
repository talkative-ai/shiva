package models

import (
	"time"
)

type AumModel struct {
	ID        *uint32    `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
}

type AumEntity struct {
	Title   string  `json:"title"`
	Created *string `json:"created_at,omitempty" db:"-"`
}

type AumProject struct {
	AumModel
	AumEntity
	OwnerID   string `json:"-"`
	StartZone uint32 `json:"startZone,omitempty"` // Expected Zone ID

	Actors []AumActor `json:"actors,omitempty" db:"-"`
	Zones  []AumZone  `json:"locations,omitempty" db:"-"`
	Notes  []AumNote  `json:"notes,omitempty" db:"-"`
}

type AumDialog struct {
	AumModel
	AumEntity
	Dialog string `json:"dialog"`
}

type AumActor struct {
	AumModel
	AumEntity

	Container         bool     `json:"container"`
	Carriable         bool     `json:"carriable"`
	ContainerContents []uint32 `json:"containerContents,omitempty"` // Expected array of Object IDs

	// TODO: Define a conversational dialog structure
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`
}

type AumZone struct {
	AumModel
	AumEntity

	Description      string                `json:"description"`
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	Objects          []uint32              `json:"objects,omitempty"`
	Actors           []uint32              `json:"actors,omitempty"`
	LinkedZones      []AumZoneLink         `json:"linkedZones,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`
}

type AumZoneLink struct {
	AumModel

	ZoneFrom     uint32
	ZoneTo       uint32
	Conditionals []AumConditional
}

type AumNote struct {
	AumModel
	AumEntity
	Text string `json:"text"`
}

type AumConditional struct {
	AumModel

	LogicalBlock   []AumComparison
	OverrideStruct interface{}
}

type AumComparison struct {
	AumModel
	Value1         interface{}
	Value2         interface{}
	LogicOperation AumLogic
}

// AumLogic specifies ducktyped logical operators
type AumLogic uint32

const (
	// AumAND is Logical AND
	AumAND AumLogic = 1 << iota
	// AumOR is Logical OR
	AumOR
	// AumEQ is equality
	AumEQ
	// AumNOT is Logical NOT
	AumNOT
	// AumGT is greater-than
	AumGT
	// AumGTE is greater-than or equal-to
	AumGTE
	// AumLT is less-than
	AumLT
	// AumLTE is less-than or equal-to
	AumLTE
)

type AumCustomProperties map[string]interface{}
