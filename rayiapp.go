package rayiapp

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kaos/deployer"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
)

// App represents the main application structure..
type App struct {
	Name string

	Config    *AppConfig
	opts      *AppOpts
	deployers map[string]DeployerConfig
	eventhubs map[string]kaos.EventServerConfig
	service   *kaos.Service
	mux       *http.ServeMux
}

// AppOpts represents the options for the application.
type AppOpts struct {
	Logger     *logger.LogEngine
	Apps       []string
	Publishers []string
}

type DeployerParamType string

const (
	DeployerParamTypeNone       DeployerParamType = "none"
	DeployerParamTypeDefaultMux DeployerParamType = "default-mux"
	DeployerParamTypeFunc       DeployerParamType = "func"
)

// DeployerConfig abstraction of deployer.Deployer
type DeployerConfig struct {
	Deployer  deployer.Deployer
	ParamType DeployerParamType
	Param     func() interface{}
}

func CreateApp(configFile string, opts *AppOpts) (*App, error) {
	if opts == nil {
		opts = &AppOpts{}
	}
	if opts.Logger == nil {
		opts.Logger = kaos.CreateLog()
	}
	app := new(App)
	app.opts = opts
	app.Name = "RayiApp-" + codekit.GenerateRandomString("ABCDEFGHIJKLMNOPQRTUVWXYZ0123456789", 6)
	app.deployers = make(map[string]DeployerConfig)
	app.eventhubs = make(map[string]kaos.EventServerConfig)
	app.service = kaos.NewService()
	app.service.SetLogger(opts.Logger)

	appConfig := NewAppConfig()
	err := ReadConfig(configFile, appConfig)
	if err != nil {
		return app, fmt.Errorf("failed to create app %v", err)
	}
	app.Config = appConfig

	for name, svcConfig := range app.Config.Services {
		if len(app.opts.Apps) > 0 && !slices.Contains(app.opts.Apps, name) {
			continue
		}
		depl, err := svcConfig.ToDeployerConfig(name, app)
		if err != nil {
			return app, fmt.Errorf("failed to create app deployer %s, %v", name, err)
		}
		app.deployers[name] = *depl
	}

	// load data connection
	if err := app.StartDataHub(); err != nil {
		return app, err
	}

	return app, nil
}

func (app *App) Start() error {
	app.mux = http.NewServeMux()
	app.opts.Logger.Infof("starting app %s", app.Name)

	// load deployer
	if err := app.StartDeployer(); err != nil {
		return err
	}

	// load publisher
	if err := app.StartPublisher(); err != nil {
		return err
	}

	return nil
}

func (app *App) StartDataHub() error {
	hm := kaos.NewHubManager(nil)
	vTenant := new(ConnectionInfo)
	vTenant.UseTx = false
	vTenant.PoolSize = 100
	vTenant.Timeout = 120
	vTenant.AutoCloseMs = 2000
	vTenant.AutoReleaseMs = 0

	for k, v := range app.Config.Connections {
		if k == "tenant" {
			vTenant.Txt = v.Txt
			vTenant.PoolSize = v.PoolSize
			vTenant.UseTx = v.UseTx
			if v.Timeout > 0 {
				vTenant.Timeout = v.Timeout
			}
			if v.AutoCloseMs > 0 {
				vTenant.AutoCloseMs = v.AutoCloseMs
			}
			if v.AutoReleaseMs > 0 {
				vTenant.AutoReleaseMs = v.AutoReleaseMs
			}
			continue
		}
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(v.Txt, v.UseTx), true, v.PoolSize)
		hconn.SetTimeout(time.Duration(v.Timeout) * time.Second)
		hconn.SetAutoCloseDuration(time.Duration(v.AutoCloseMs) * time.Millisecond)
		hconn.SetAutoReleaseDuration(time.Duration(v.AutoReleaseMs) * time.Millisecond)
		hm.Set(k, "", hconn)
		app.Logger().Infof("loading data connection %s", k)
	}

	hm.SetHubBuilder(func(key, group string) (*datahub.Hub, error) {
		vTenantConnStr := vTenant.Txt
		if strings.Contains(vTenantConnStr, "%s") {
			vTenantConnStr = fmt.Sprintf(vTenantConnStr, key)
		}
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(vTenantConnStr, vTenant.UseTx), true, vTenant.PoolSize)
		hconn.SetAutoCloseDuration(2 * time.Second)
		hconn.SetAutoReleaseDuration(0 * time.Second)
		return hconn, nil
	})

	app.service.SetHubManager(hm)
	return nil
}

func (app *App) StartDeployer() error {
	for name, ad := range app.deployers {
		if err := app.deploy(ad); err != nil {
			return fmt.Errorf("failed to deploy %s. %v", name, err)
		}
	}
	return nil
}

func (app *App) StartPublisher() error {
	for name, svcConfig := range app.Config.Services {
		if len(app.opts.Publishers) > 0 {
			if !slices.Contains(app.opts.Publishers, name) {
				continue
			}
		}
		if svcConfig.ServerType != string(kaos.ServiceTypeEvent) {
			continue
		}
		svcConfig.Secret = Update(svcConfig.Secret)
		svcConfig.Signature = Update(svcConfig.Signature)

		for k, v := range svcConfig.Data {
			if vs, ok := v.(string); ok {
				svcConfig.Data[k] = Update(vs)
			}
		}

		evConfig := svcConfig.ToEventServerConfig(name, app)
		ev, err := kaos.NewEventHub(*evConfig)
		basePath := svcConfig.Data.Get("base_path", "").(string)
		if basePath != "" {
			ev.SetPrefix(basePath)
		}
		if err != nil {
			return fmt.Errorf("failed to load publisher for event hub %s. %v", name, err)
		}
		app.service.RegisterEventHub(ev, name, svcConfig.Secret)
		app.Logger().Infof("loading publisher for %s, %s, %s", name, svcConfig.Provider, evConfig.Server)
	}
	return nil
}

func (app *App) deploy(ad DeployerConfig) error {
	app.Logger().Infof("deploying %s, %s, %v", ad.Deployer.Name(), ad.ParamType, ad.Deployer.Get("host"))
	if ad.Deployer == nil {
		return app.Logger().Error("deployer is nil")
	}
	if ad.ParamType == DeployerParamTypeNone {
		return ad.Deployer.Deploy(app.service, nil)
	} else if ad.ParamType == DeployerParamTypeDefaultMux {
		return ad.Deployer.Deploy(app.service, app.mux)
	}

	param := ad.Param()
	return ad.Deployer.Deploy(app.service, param)
}

func (app *App) WaitForGraceShutdown() {
	// grace shutdown
	csign := make(chan os.Signal, 1)
	signal.Notify(csign, os.Interrupt)
	<-csign
	app.Logger().Infof("shutting down app %s", app.Name)
}

func (app *App) Exit(code int, message string) {
	if code >= 0 {
		app.Logger().Info(message)
		os.Exit(code)
	}

	if code < 0 {
		app.Logger().Error(message)
		os.Exit(code)
	}
}

func (app *App) Logger() *logger.LogEngine {
	return app.opts.Logger
}

func (app *App) RegisterDeployer(appDeployer DeployerConfig) {
	if appDeployer.Deployer == nil {
		return
	}
	if app.deployers == nil {
		app.deployers = make(map[string]DeployerConfig)
	}
	app.deployers[appDeployer.Deployer.Name()] = appDeployer
}

func (app *App) RegisterEventHub(name string, hub kaos.EventServerConfig) {
	if app.eventhubs == nil {
		app.eventhubs = make(map[string]kaos.EventServerConfig)
	}
	app.eventhubs[name] = hub
}

func (app *App) GetHost(name string) string {
	host, ok := app.Config.Hosts[name]
	if !ok {
		return ""
	}
	return host
}

func (app *App) Service() *kaos.Service {
	return app.service
}

func (app *App) EventConfig(name string) ServiceConfig {
	svcCfg, ok := app.Config.Services[name]
	if !ok {
		return ServiceConfig{}
	}
	return svcCfg
}
