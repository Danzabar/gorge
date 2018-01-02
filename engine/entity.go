package engine

import (
    "time"
)

type (

    // Entity represents a saveable object
    Entity struct {
        ID        uint       `gorm:"primary_key" json:"id"`
        CreatedAt time.Time  `json:"createdAt"`
        UpdatedAt time.Time  `json:"updatedAt"`
        DeletedAt *time.Time `json:"-"`
    }
)
