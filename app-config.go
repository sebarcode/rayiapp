package rayiapp

import (
	"os"
	"strings"

	"github.com/sebarcode/codekit"
)

var (
	AppConfigVariables codekit.M
)

type ConnectionInfo struct {
	Txt           string
	UseTx         bool
	PoolSize      int
	Timeout       int
	AutoReleaseMs int
	AutoCloseMs   int
}

type AppConfig struct {
	Hosts       map[string]string
	Connections map[string]ConnectionInfo

	Services map[string]ServiceConfig
	Data     codekit.M
}

func NewAppConfig() *AppConfig {
	a := new(AppConfig)
	a.Hosts = make(map[string]string)
	a.Connections = make(map[string]ConnectionInfo)
	return a
}

func (cfg *AppConfig) Parse() {
	for id, host := range cfg.Hosts {
		cfg.Hosts[id] = Update(host)
	}

	for id, data := range cfg.Data {
		dataStr, ok := data.(string)
		if ok {
			cfg.Data[id] = Update(dataStr)
		}
	}

	for id, conn := range cfg.Connections {
		conn.Txt = Update(conn.Txt)
		cfg.Connections[id] = conn
	}

	for _, ev := range cfg.Services {
		ev.Server = Update(ev.Server)
		ev.Secret = Update(ev.Secret)
		ev.Signature = Update(ev.Signature)
	}
}

func (cfg *AppConfig) DataToEnv() {
	for k, v := range cfg.Data {
		switch value := v.(type) {
		case string:
			os.Setenv(k, value)
		}
	}
}

func UpdateWithEnv(txt string) string {
	parts := strings.Split(txt, "${env:")
	if len(parts) <= 1 {
		return txt
	}

	for _, part := range parts[1:] {
		envID := strings.Split(part, "}")[0]
		envValue := os.Getenv(envID)
		txt = strings.ReplaceAll(txt, "${env:"+envID+"}", envValue)
	}

	return txt
}

func UpdateWithVar(txt string) string {
	parts := strings.Split(txt, "${var:")
	if len(parts) <= 1 {
		return txt
	}

	for _, part := range parts[1:] {
		id := strings.Split(part, "}")[0]
		value, ok := AppConfigVariables[id].(string)
		if ok {
			txt = strings.ReplaceAll(txt, "${var:"+id+"}", value)
		}
	}

	return txt
}

func UpdateWithVarFromM(txt string, m codekit.M) string {
	parts := strings.Split(txt, "${var:")
	if len(parts) <= 1 {
		return txt
	}

	for _, part := range parts[1:] {
		id := strings.Split(part, "}")[0]
		value, ok := m[id].(string)
		if ok {
			txt = strings.ReplaceAll(txt, "${var:"+id+"}", value)
		}
	}
	return txt
}

func Update(txt string) string {
	s := UpdateWithEnv(txt)
	if s == "" || s == txt {
		s = UpdateWithVar(txt)
	}

	wd, e := os.Getwd()
	if e == nil && strings.Contains(s, "${ctx:wd}") {
		s = strings.ReplaceAll(s, "${ctx:wd}", wd)
	}

	return s
}
