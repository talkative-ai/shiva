package models

type AumID int64

type AumProject struct {
	ID      int64  `json:"id" datastore:"-"`
	Title   string `json:"title"`
	OwnerID string `json:"-"`

	NPCs      []AumNPC      `json:"npcs,omitempty" datastore:"-"`
	Objects   []AumObject   `json:"objects,omitempty" datastore:"-"`
	Locations []AumLocation `json:"locations,omitempty" datastore:"-"`
	Notes     []AumNote     `json:"notes,omitempty" datastore:"-"`
}

type AumNPC struct {
	ID    int64 `json:"id" datastore:"-"`
	Title string
	// TODO: Define a conversational dialog structure
	CustomProperties []AumCustomProperties
	Conditionals     []AumConditional
}

type AumObject struct {
	ID                int64 `json:"id" datastore:"-"`
	Title             string
	Container         bool
	Carriable         bool
	ContainerContents []int64
	Conditionals      []AumConditional
	CustomProperties  []AumCustomProperties
}

type AumLocation struct {
	ID               int64                 `json:"id" datastore:"-"`
	Title            string                `json:"title"`
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	Objects          []int64               `json:"objects,omitempty"`
	NPCs             []int64               `json:"npcs,omitempty"`
	LinkedLocations  []AumLocationLink     `json:"linked_locations,omitempty"`
	CustomProperties []AumCustomProperties `json:"custom_properties,omitempty"`
}

type AumLocationLink struct {
	LocationFrom int64
	LocationTo   int64
	Conditionals []AumConditional
}

type AumNote struct {
	ID    int64 `json:"id" datastore:"-"`
	Title string
	Text  string
}

type AumConditional struct {
	LogicalBlock   []AumComparison
	OverrideStruct interface{}
}

type AumComparison struct {
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
