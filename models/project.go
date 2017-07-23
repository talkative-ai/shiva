package models

import (
	"database/sql"
	"time"
)

type AumModel struct {
	ID        *uint64    `json:"id" db:"id, primarykey, autoincrement"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
}

type AumEntity struct {
	Title   string  `json:"title" db:"title"`
	Created *string `json:"created_at,omitempty" db:"-"`
}

type AumProject struct {
	AumModel
	AumEntity
	OwnerID   string        `json:"-" db:"owner_id"`
	StartZone sql.NullInt64 `json:"startZone,omitempty" db:"start_zone_id"` // Expected Zone ID

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
	ContainerContents []uint64 `json:"containerContents,omitempty"` // Expected array of Object IDs

	// TODO: Define a conversational dialog structure
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`
}

type AumZone struct {
	AumModel
	AumEntity

	Description      string                `json:"description"`
	Objects          []uint64              `json:"objects,omitempty"`
	Actors           []uint64              `json:"actors,omitempty"`
	LinkedZones      []AumZoneLink         `json:"linkedZones,omitempty"`
	CustomProperties []AumCustomProperties `json:"customProperties,omitempty"`
}

type AumZoneLink struct {
	AumModel

	ZoneFrom uint64
	ZoneTo   uint64
}

type AumNote struct {
	AumModel
	AumEntity
	Text string `json:"text"`
}

type AumCustomProperties map[string]interface{}
