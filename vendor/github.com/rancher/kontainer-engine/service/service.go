package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rancher/kontainer-engine/cluster"
	"github.com/rancher/kontainer-engine/drivers"
	"github.com/rancher/kontainer-engine/types"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	pluginAddress = map[string]string{}
	Drivers       = map[string]types.Driver{}
)

const (
	ListenAddress = "127.0.0.1:"
)

func Start() error {
	for name, driver := range drivers.Drivers {
		RegisterDriver(name, driver)
	}
	return nil
}

func RegisterDriver(name string, driver types.Driver) {
	RegisterDriverForPort(name, driver, 0)
}

func RegisterDriverForPort(name string, driver types.Driver, port int) {
	logrus.Infof("Activating driver %s", name)
	addr := make(chan string)
	if port == 0 {
		go types.NewServer(driver, addr).Serve(ListenAddress)
	} else {
		go types.NewServer(driver, addr).Serve(ListenAddress + strconv.Itoa(port))
	}

	Drivers[name] = driver
	listenAddr := <-addr
	pluginAddress[name] = listenAddr
	logrus.Infof("Activating driver %s done", name)
}

func RegisterExternalDriver(name string, address string) {
	pluginAddress[name] = address
}

func DeregisterExternalDriver(name string) {
	delete(pluginAddress, name)
}

type controllerConfigGetter struct {
	driverName  string
	clusterSpec v3.ClusterSpec
	clusterName string
}

func (c controllerConfigGetter) GetConfig() (types.DriverOptions, error) {
	driverOptions := types.DriverOptions{
		BoolOptions:        make(map[string]bool),
		StringOptions:      make(map[string]string),
		IntOptions:         make(map[string]int64),
		StringSliceOptions: make(map[string]*types.StringSlice),
	}
	data := map[string]interface{}{}
	switch c.driverName {
	case "import":
		config, err := toMap(c.clusterSpec.ImportedConfig, "json")
		if err != nil {
			return driverOptions, err
		}
		data = config
		flatten(data, &driverOptions)
	default:
		config, err := toMap(c.clusterSpec.GenericEngineConfig, "json")
		if err != nil {
			return driverOptions, err
		}
		data = config
		flatten(data, &driverOptions)
	}

	driverOptions.StringOptions["name"] = c.clusterName
	displayName := c.clusterSpec.DisplayName
	if displayName == "" {
		displayName = c.clusterName
	}
	driverOptions.StringOptions["displayName"] = displayName

	return driverOptions, nil
}

// flatten take a map and flatten it and convert it into driverOptions
func flatten(data map[string]interface{}, driverOptions *types.DriverOptions) {
	for k, v := range data {
		switch v.(type) {
		case float64:
			driverOptions.IntOptions[k] = int64(v.(float64))
		case string:
			driverOptions.StringOptions[k] = v.(string)
		case bool:
			driverOptions.BoolOptions[k] = v.(bool)
		case []interface{}:
			// lists of strings come across as lists of interfaces, have to convert them manually
			var stringArray []string

			for _, stringInterface := range v.([]interface{}) {
				switch stringInterface.(type) {
				case string:
					stringArray = append(stringArray, stringInterface.(string))
				}
			}

			// if the length is 0 then it must not have been an array of strings
			if len(stringArray) != 0 {
				driverOptions.StringSliceOptions[k] = &types.StringSlice{Value: stringArray}
			}
		case []string:
			driverOptions.StringSliceOptions[k] = &types.StringSlice{Value: v.([]string)}
		case map[string]interface{}:
			// hack for labels
			if k == "labels" {
				r := []string{}
				for key1, value1 := range v.(map[string]interface{}) {
					r = append(r, fmt.Sprintf("%v=%v", key1, value1))
				}
				driverOptions.StringSliceOptions[k] = &types.StringSlice{Value: r}
			} else {
				flatten(v.(map[string]interface{}), driverOptions)
			}
		default:
			logrus.Warnf("could not convert %v %v=%v", reflect.TypeOf(v), k, v)
		}
	}
}

func toMap(obj interface{}, format string) (map[string]interface{}, error) {
	if format == "json" {
		data, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	} else if format == "yaml" {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		var result map[string]interface{}
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, nil
}

type EngineService interface {
	Create(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (string, string, string, error)
	Update(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (string, string, string, error)
	Remove(ctx context.Context, name string, clusterSpec v3.ClusterSpec) error
	GetDriverCreateOptions(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (*types.DriverFlags, error)
	GetDriverUpdateOptions(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (*types.DriverFlags, error)
}

type engineService struct {
	store cluster.PersistentStore
}

func NewEngineService(store cluster.PersistentStore) EngineService {
	return &engineService{
		store: store,
	}
}

func (e *engineService) convertCluster(name string, spec v3.ClusterSpec) (cluster.Cluster, error) {
	// todo: decide whether we need a driver field
	driverName := ""
	if spec.ImportedConfig != nil {
		driverName = "import"
	} else if spec.GenericEngineConfig != nil {
		driverName = (*spec.GenericEngineConfig)["driverName"].(string)
		if driverName == "" {
			return cluster.Cluster{}, fmt.Errorf("no driver name supplied")
		}
	}
	if driverName == "" {
		return cluster.Cluster{}, fmt.Errorf("no driver config found")
	}
	pluginAddr := pluginAddress[driverName]
	configGetter := controllerConfigGetter{
		driverName:  driverName,
		clusterSpec: spec,
		clusterName: name,
	}
	clusterPlugin, err := cluster.NewCluster(driverName, pluginAddr, name, configGetter, e.store)
	if err != nil {
		return cluster.Cluster{}, err
	}
	return *clusterPlugin, nil
}

// Create creates the stub for cluster manager to call
func (e *engineService) Create(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (string, string, string, error) {
	cls, err := e.convertCluster(name, clusterSpec)
	if err != nil {
		return "", "", "", err
	}
	if err := cls.Create(ctx); err != nil {
		return "", "", "", err
	}
	endpoint := cls.Endpoint
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = fmt.Sprintf("https://%s", cls.Endpoint)
	}
	return endpoint, cls.ServiceAccountToken, cls.RootCACert, nil
}

// Update creates the stub for cluster manager to call
func (e *engineService) Update(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (string, string, string, error) {
	cls, err := e.convertCluster(name, clusterSpec)
	if err != nil {
		return "", "", "", err
	}
	if err := cls.Update(ctx); err != nil {
		return "", "", "", err
	}
	endpoint := cls.Endpoint
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = fmt.Sprintf("https://%s", cls.Endpoint)
	}
	return endpoint, cls.ServiceAccountToken, cls.RootCACert, nil
}

// Remove removes stub for cluster manager to call
func (e *engineService) Remove(ctx context.Context, name string, clusterSpec v3.ClusterSpec) error {
	cls, err := e.convertCluster(name, clusterSpec)
	if err != nil {
		return err
	}
	return cls.Remove(ctx)
}

func (e *engineService) GetDriverCreateOptions(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (*types.DriverFlags,
	error) {
	cls, err := e.convertCluster(name, clusterSpec)
	if err != nil {
		return nil, err
	}

	return cls.GetDriverCreateOptions(ctx)
}

func (e *engineService) GetDriverUpdateOptions(ctx context.Context, name string, clusterSpec v3.ClusterSpec) (*types.DriverFlags,
	error) {
	cls, err := e.convertCluster(name, clusterSpec)
	if err != nil {
		return nil, err
	}

	return cls.GetDriverUpdateOptions(ctx)
}
