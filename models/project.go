package models

type AumID int64

type AumProject struct {
	ID      int64  `json:"id" datastore:"-"`
	Title   string `json:"title"`
	OwnerID string `json:"-"`

	NPCs      []AumNPC      `json:"npcs" datastore:"-"`
	Objects   []AumObject   `json:"objects" datastore:"-"`
	Locations []AumLocation `json:"locations" datastore:"-"`
	Notes     []AumNote     `json:"notes" datastore:"-"`
}

type AumNPC struct {
	ID    int64 `json:"id" datastore:"-"`
	Name  string
	Title string
	// TODO: Define a conversational dialog structure
	CustomProperties []AumCustomProperties
	Conditionals     []AumConditional
}

type AumObject struct {
	ID                int64 `json:"id" datastore:"-"`
	Name              string
	Title             string
	Container         bool
	Carriable         bool
	ContainerContents []int64
	Conditionals      []AumConditional
	CustomProperties  []AumCustomProperties
}

type AumLocation struct {
	ID               int64                 `json:"id" datastore:"-"`
	Name             string                `json:"name"`
	Title            string                `json:"title"`
	Conditionals     []AumConditional      `json:"conditionals"`
	Objects          []int64               `json:"objects"`
	NPCs             []int64               `json:"npcs"`
	LinkedLocations  []AumLocationLink     `json:"linked_locations"`
	CustomProperties []AumCustomProperties `json:"custom_properties"`
}

type AumLocationLink struct {
	LocationFrom int64
	LocationTo   int64
	Conditionals []AumConditional
}

type AumNote struct {
	ID    int64 `json:"id" datastore:"-"`
	Name  string
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
	AumOR AumLogic = 1 << iota
	// AumEQ is equality
	AumEQ AumLogic = 1 << iota
	// AumNOT is Logical NOT
	AumNOT AumLogic = 1 << iota
	// AumGT is greater-than
	AumGT AumLogic = 1 << iota
	// AumGTE is greater-than or equal-to
	AumGTE AumLogic = 1 << iota
	// AumLT is less-than
	AumLT AumLogic = 1 << iota
	// AumLTE is less-than or equal-to
	AumLTE AumLogic = 1 << iota
)

type AumCustomProperties map[string]interface{}
