package models

import (
	"time"
)

type AumID int64

type AumModel struct {
	ID        *uint64   `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type AumEntity struct {
	Title   string  `json:"title"`
	Created *string `json:"created,omitempty" db:"-"`
}

type AumProject struct {
	AumModel
	OwnerID       string `json:"-"`
	StartLocation int64  `json:"startLocation,omitempty"` // Expected Location ID

	NPCs      []AumNPC      `json:"npcs,omitempty" db:"-"`
	Objects   []AumObject   `json:"objects,omitempty" db:"-"`
	Locations []AumLocation `json:"locations,omitempty" db:"-"`
	Notes     []AumNote     `json:"notes,omitempty" db:"-"`
}

type AumDialogue struct {
	AumModel
	AumEntity
	Dialog string `json:"dialog"`
}

type AumNPC struct {
	AumModel
	AumEntity

	// TODO: Define a conversational dialog structure
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`

	Created *string `json:"created,omitempty" db:"-"`
}

type AumObject struct {
	AumModel
	AumEntity

	Container         bool                  `json:"container"`
	Carriable         bool                  `json:"carriable"`
	Locations         []int64               `json:"locations,omitempty"`         // Expected array of Location IDs
	ContainerContents []int64               `json:"containerContents,omitempty"` // Expected array of Object IDs
	Conditionals      []AumConditional      `json:"conditionals,omitempty"`
	CustomProperties  []AumCustomProperties `json:"customProperties,omitempty"`
}

type AumLocation struct {
	AumModel
	AumEntity

	Description      string                `json:"description"`
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	Objects          []int64               `json:"objects,omitempty"`
	NPCs             []int64               `json:"npcs,omitempty"`
	LinkedLocations  []AumLocationLink     `json:"linkedLocations,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`

	Created *string `json:"created,omitempty" db:"-"`
}

type AumLocationLink struct {
	AumModel

	LocationFrom int64
	LocationTo   int64
	Conditionals []AumConditional
}

type AumNote struct {
	AumModel
	AumEntity
	Text string `json:"text"`

	Created *string `json:"created,omitempty" db:"-"`
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
type AumLogic uint64

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
