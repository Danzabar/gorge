package engine

import (
    "errors"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "strings"
    "sync"

    "gopkg.in/yaml.v2"
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
        config  *sync.Map
    }

    // Config represents a single element of config, these are stored
    // with their raw values and can be later converted to a struct
    // when requested
    Config struct {
        Raw  []byte `json:"raw"`
        Name string `json:"name"`
    }

    // GorgeSettings are used to control engine behaivour
    // and may be used to incorperate other tools in the future
    GorgeSettings struct {
        Game   GorgeSettingsDetail `yaml:"game"`
        Config []string            `yaml:"config"`
    }

    // GorgeSettingsDetail defines the main game details
    // these are more for display, unless you build logic around
    // them
    GorgeSettingsDetail struct {
        Name    string `yaml:"name"`
        Version string `yaml:"version"`
    }
)

func NewConfig(GM *GameManager) *ConfigManager {
    return &ConfigManager{
        gm:     GM,
        config: new(sync.Map),
    }
}

func (c *ConfigManager) AddTarget(n ...string) {
    for _, v := range n {
        c.targets = append(c.targets, v)
    }
}

func (c *ConfigManager) LoadStandard() {
    c.Fetch(STANDARD_CONFIG)

    // Convert it
    var st GorgeSettings

    if err := c.ConvertYaml("gorge", &st); err != nil {
        c.gm.Log.Error("Unable to convert standard config: " + err.Error())
        return
    }

    // Otherwise load the settings into the GM
    c.gm.Settings = &st

    // If we have values for config, use them
    if len(st.Config) > 0 {
        c.AddTarget(st.Config...)
    }
}

func (c *ConfigManager) Load() {
    for _, v := range c.targets {
        c.gm.Log.Debug(v)
        // is it a file or directory?
        fi, err := os.Stat(v)

        if err != nil {
            c.gm.Log.Error(err)
            continue
        }

        // If its a file and not a directory, we can just
        // load this
        if !fi.IsDir() {
            c.Fetch(v)
            continue
        }

        // Otherwise we need to traverse a directory
        go c.traverseDir(v)
    }
}

func (c *ConfigManager) ConvertYaml(n string, i interface{}) error {
    // Fetch config
    r, k := c.config.Load(n)

    if !k {
        c.gm.Log.Warningf("Trying to convert a configuration file that doesn't exist: %s", n)
        return errors.New("no config found for name: " + n)
    }

    // Convert and return
    con := r.(Config)

    return yaml.Unmarshal(con.Raw, i)
}

func (c *ConfigManager) Fetch(n string) {
    var con Config

    c.gm.Log.Debugf("Loading configuration file: %s", n)

    data, err := ioutil.ReadFile(n)

    if err != nil {
        c.gm.Log.Error(err)
        return
    }

    b := path.Base(n)
    name := strings.TrimSuffix(b, filepath.Ext(b))

    con = Config{Raw: data, Name: name}
    c.config.Store(con.Name, con)

    c.gm.Log.Debug(con.Name)
}

func (c *ConfigManager) traverseDir(n string) {
    err := filepath.Walk(n, func(path string, f os.FileInfo, err error) error {
        if f.IsDir() {
            return nil
        }

        c.Fetch(path)
        return nil
    })

    if err != nil {
        c.gm.Log.Error(err)
    }
}
