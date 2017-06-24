package models

type AumID int64

type AumProject struct {
	ID            int64  `json:"id" datastore:"-"`
	Title         string `json:"title"`
	OwnerID       string `json:"-"`
	StartLocation int64  `json:"startLocation,omitempty"` // Expected Location ID

	NPCs      []AumNPC      `json:"npcs,omitempty" datastore:"-"`
	Objects   []AumObject   `json:"objects,omitempty" datastore:"-"`
	Locations []AumLocation `json:"locations,omitempty" datastore:"-"`
	Notes     []AumNote     `json:"notes,omitempty" datastore:"-"`
}

type AumDialogue struct {
	ID     *int64 `json:"id"`
	Title  string `json:"title"`
	Dialog string `json:"dialog"`
}

type AumNPC struct {
	ID    *int64 `json:"id" datastore:"-"`
	Title string `json:"title"`
	// TODO: Define a conversational dialog structure
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`

	Created *string `json:"created,omitempty" datastore:"-"`
}

type AumObject struct {
	ID                *int64                `json:"id" datastore:"-"`
	Title             string                `json:"title"`
	Container         bool                  `json:"container"`
	Carriable         bool                  `json:"carriable"`
	Locations         []int64               `json:"locations,omitempty"`         // Expected array of Location IDs
	ContainerContents []int64               `json:"containerContents,omitempty"` // Expected array of Object IDs
	Conditionals      []AumConditional      `json:"conditionals,omitempty"`
	CustomProperties  []AumCustomProperties `json:"customProperties,omitempty"`

	Created *string `json:"created,omitempty" datastore:"-"`
}

type AumLocation struct {
	ID               *int64                `json:"id" datastore:"-"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	Conditionals     []AumConditional      `json:"conditionals,omitempty"`
	Objects          []int64               `json:"objects,omitempty"`
	NPCs             []int64               `json:"npcs,omitempty"`
	LinkedLocations  []AumLocationLink     `json:"linkedLocations,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`

	Created *string `json:"created,omitempty" datastore:"-"`
}

type AumLocationLink struct {
	LocationFrom int64
	LocationTo   int64
	Conditionals []AumConditional
}

type AumNote struct {
	ID    *int64 `json:"id" datastore:"-"`
	Title string `json:"title"`
	Text  string `json:"text"`

	Created *string `json:"created,omitempty" datastore:"-"`
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
