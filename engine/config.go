package engine

import (
    "path/filepath"
)

const (
    // Constant value for the location of the standard
    // configuration yaml
    STANDARD_CONFIG = "./gorge.yaml"
)

type (
    // ConfigManager is responsible for fetching and displaying config
    // information, there are two types of configuration, the standard
    // gorge.yaml and end user provided directories
    ConfigManager struct {
        gm      *GameManager
        targets []string
    }

    // GorgeSettings are used to control engine behaivour
    // and may be used to incorperate other tools in the future
    GorgeSettings struct {
        Game GorgeSettingsDetail `yaml:"game"`
    }

    // GorgeSettingsDetail defines the main game details
    // these are more for display, unless you build logic around
    // them
    GorgeSettingsDetail struct {
        Name    string `yaml:"name"`
        Version string `yaml:"version"`
    }
)
