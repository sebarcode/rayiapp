package rayiapp

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kaos/deployer"
	"github.com/sebarcode/codekit"
)

type ServiceConfig struct {
	Name       string
	Server     string
	ServerType string
	Provider   string
	ParamType  DeployerParamType
	Timeout    int
	Secret     string
	Signature  string
	ByterName  string
	Data       codekit.M
}

func (s *ServiceConfig) ToEventServerConfig(serviceName string, app *App) *kaos.EventServerConfig {
	host := app.GetHost(serviceName)
	if s.Data == nil {
		s.Data = codekit.M{}
	}
	return &kaos.EventServerConfig{
		Name:              serviceName,
		Server:            host,
		Provider:          s.Provider,
		DeployerParamType: kaos.DeployerParamType(s.ParamType),
		Timeout:           s.Timeout,
		Secret:            s.Secret,
		Signature:         s.Signature,
		ByterName:         s.ByterName,
		Data:              s.Data.Set("host", host),
	}
}

func (s *ServiceConfig) ToDeployerConfig(serviceName string, app *App) (*DeployerConfig, error) {
	var depl deployer.Deployer
	if s.Data == nil {
		s.Data = codekit.M{}
	}
	if s.Data.GetBool("require_validation") {
		s.Data.Set("secret", s.Secret)
	}
	depl, err := deployer.GetDeployer(s.Provider, s.Data)
	if err != nil {
		return nil, err
	}

	depl.Set("host", app.GetHost(serviceName))
	for k, v := range s.Data {
		depl.Set(k, v)
	}

	var deplParam func() interface{}

	switch s.ParamType {
	case DeployerParamTypeDefaultMux:
		deplParam = nil

	case DeployerParamTypeFunc:
		deplParam = depl.DefaultDeployerParam()
	}

	return &DeployerConfig{
		Deployer:  depl,
		ParamType: s.ParamType,
		Param:     deplParam,
	}, nil
}
