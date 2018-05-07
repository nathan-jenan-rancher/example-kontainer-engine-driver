package main

import (
	"context"

	"github.com/rancher/kontainer-engine/types"
	"github.com/sirupsen/logrus"
)

type MyDriver struct {
	types.UnimplementedClusterSizeAccess
	types.UnimplementedVersionAccess
}

func (m *MyDriver) GetDriverCreateOptions(ctx context.Context) (*types.DriverFlags, error) {
	driverFlag := types.DriverFlags{
		Options: make(map[string]*types.Flag),
	}
	driverFlag.Options["name"] = &types.Flag{
		Type:  types.StringType,
		Usage: "The internal name of the cluster in Rancher",
	}
	return &driverFlag, nil
}

func (m *MyDriver) GetDriverUpdateOptions(ctx context.Context) (*types.DriverFlags, error) {
	return nil, nil
}

func (m *MyDriver) Create(ctx context.Context, opts *types.DriverOptions, clusterInfo *types.ClusterInfo) (*types.ClusterInfo, error) {
	logrus.Infof("mydriver create called")
	logrus.Infof("options provided: %v", opts)
	logrus.Infof("cluster info: %v", clusterInfo)
	return &types.ClusterInfo{}, nil
}

func (m *MyDriver) Update(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions) (*types.ClusterInfo, error) {
	logrus.Infof("mydriver updated called")
	return clusterInfo, nil
}

func (m *MyDriver) PostCheck(ctx context.Context, clusterInfo *types.ClusterInfo) (*types.ClusterInfo, error) {
	logrus.Infof("mydriver post check called")
	return clusterInfo, nil
}

func (m *MyDriver) Remove(ctx context.Context, clusterInfo *types.ClusterInfo) error {
	logrus.Infof("mydriver remove called")
	return nil
}

func (m *MyDriver) GetCapabilities(ctx context.Context) (*types.Capabilities, error) {
	logrus.Infof("mydriver getcaps called")
	return nil, nil
}
