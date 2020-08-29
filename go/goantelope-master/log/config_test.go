package log

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

var (
	confData = []struct {
		ok          bool
		cfg         *Config
		jsonContent string
		yamlContent string
	}{
		{
			ok: true,
			cfg: &Config{
				Level:         "info",
				Path:          "/var/www/application.log",
				Development:   false,
				FlushCount:    2,
				FlushInterval: 2,
			},
			jsonContent: `
{
    "level": "info",
    "path": "/var/www/application.log",
    "development": false,
	"flush_count": 2,
	"flush_interval": 2
}
			`,
			yamlContent: `
level: info
path: /var/www/application.log
development: false
flush_count: 2
flush_interval: 2`,
		},
		{
			ok: true,
			cfg: &Config{
				Level:         "debug",
				Path:          "/var/www/application2.log",
				Development:   true,
				FlushCount:    defaultFlushCount,
				FlushInterval: defaultFlushInterval,
			},
			jsonContent: `
{
    "level": "debug",
    "path": "/var/www/application2.log",
    "development": true
}
			`,
			yamlContent: `
level: debug
path: /var/www/application2.log
development: true`,
		},
		{
			ok: false,
			cfg: &Config{
				Level:       "debug",
				Path:        "",
				Development: true,
			},
			jsonContent: `
{
    "level": "debug",
    "development": true
}
			`,
			yamlContent: `
level: debug
development: true`,
		},
		{
			ok: true,
			cfg: &Config{
				Level:         "info",
				Path:          "/var/www/application3.log",
				Development:   true,
				FlushCount:    defaultFlushCount,
				FlushInterval: defaultFlushInterval,
			},
			jsonContent: `
{
    "level": "xxx",
	"path": "/var/www/application3.log",
    "development": true
}
			`,
			yamlContent: `
level: xxx
path: /var/www/application3.log
development: true`,
		},
		{
			ok: true,
			cfg: &Config{
				Level:         "info",
				Path:          "/var/www/application4.log",
				Development:   false,
				FlushCount:    defaultFlushCount,
				FlushInterval: defaultFlushInterval,
			},
			jsonContent: `
{
	"path": "/var/www/application4.log",
    "development": false
}
			`,
			yamlContent: `
path: /var/www/application4.log
development: false`,
		},
	}
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	for _, data := range confData {
		jsonCfg := &Config{}
		yamlCfg := &Config{}

		err := json.Unmarshal([]byte(data.jsonContent), jsonCfg)
		assert.Nil(err)
		err = yaml.Unmarshal([]byte(data.yamlContent), yamlCfg)
		assert.Nil(err)

		assert.Equal(true, testUtilConfigEqual(jsonCfg, yamlCfg))

		cfg := yamlCfg

		assert.Equal(data.ok, cfg.Validate())

		assert.Equal(data.cfg.Path, cfg.Path)
		assert.Equal(data.cfg.Development, cfg.Development)

		zapCfg := cfg.zapCfg()
		assert.Equal(data.cfg.Level, zapCfg.Level.String())
	}
}

func testUtilConfigEqual(a, b *Config) bool {
	if a.Level != b.Level {
		return false
	}
	if a.Path != b.Path {
		return false
	}
	if a.Development != b.Development {
		return false
	}
	if a.FlushCount != b.FlushCount {
		return false
	}
	if a.FlushInterval != b.FlushInterval {
		return false
	}
	return true
}
