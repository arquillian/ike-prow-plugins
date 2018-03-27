package assets

import (
	"github.com/arquillian/ike-prow-plugins/pkg/assets/generated"
	"github.com/arquillian/ike-prow-plugins/pkg/config"
)

//go:generate go-bindata -prefix "./config/" -pkg assets -o ./generated/assets_config.go ./config/...

// LocalLoadableConfig holds a name of a config file to be loaded
type LocalLoadableConfig struct {
	ConfigFileName string
}

// Sources loads local config file that is located in pkg/assets/config directory
func (i *LocalLoadableConfig) Sources() []config.Source {
	return []config.Source{func() ([]byte, error) {
		file, err := assets.Asset(i.ConfigFileName)
		if err != nil {
			return nil, err
		}
		return file, nil
	}}
}
