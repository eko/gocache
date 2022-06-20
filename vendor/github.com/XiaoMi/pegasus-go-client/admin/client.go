package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/admin"
	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/session"
)

// Client provides the administration API to a specific cluster.
// Remember only the superusers configured to the cluster have the admin priviledges.
type Client interface {
	CreateTable(ctx context.Context, tableName string, partitionCount int) error

	DropTable(ctx context.Context, tableName string) error

	ListTables(ctx context.Context) ([]*TableInfo, error)
}

// TableInfo is the table information.
type TableInfo struct {
	Name string

	// Envs is a set of attributes binding to this table.
	Envs map[string]string
}

type Config struct {
	MetaServers []string `json:"meta_servers"`
}

// NewClient returns an instance of Client.
func NewClient(cfg Config) Client {
	return &rpcBasedClient{
		metaManager: session.NewMetaManager(cfg.MetaServers, session.NewNodeSession),
	}
}

type rpcBasedClient struct {
	metaManager *session.MetaManager
}

func (c *rpcBasedClient) waitTableReady(ctx context.Context, tableName string, partitionCount int) error {
	const replicaCount int = 3

	for {
		resp, err := c.metaManager.QueryConfig(ctx, tableName)
		if err != nil {
			return err
		}
		if resp.GetErr().Errno != base.ERR_OK.String() {
			return fmt.Errorf("QueryConfig failed: %s", resp.GetErr().String())
		}

		readyCount := 0
		for _, part := range resp.Partitions {
			if part.Primary.GetRawAddress() != 0 && len(part.Secondaries)+1 == replicaCount {
				readyCount++
			}
		}
		if readyCount == partitionCount {
			break
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (c *rpcBasedClient) CreateTable(ctx context.Context, tableName string, partitionCount int) error {
	_, err := c.metaManager.CreateApp(ctx, &admin.CreateAppRequest{
		AppName: tableName,
		Options: &admin.CreateAppOptions{
			PartitionCount: int32(partitionCount),
			ReplicaCount:   3,
			AppType:        "pegasus",
			Envs:           make(map[string]string),
			IsStateful:     true,
		},
	})
	if err != nil {
		return err
	}
	err = c.waitTableReady(ctx, tableName, partitionCount)
	return err
}

func (c *rpcBasedClient) DropTable(ctx context.Context, tableName string) error {
	req := admin.NewDropAppRequest()
	req.AppName = tableName
	reserveSeconds := int64(1) // delete immediately. the caller is responsible for the soft deletion of table.
	req.Options = &admin.DropAppOptions{
		SuccessIfNotExist: true,
		ReserveSeconds:    &reserveSeconds,
	}
	_, err := c.metaManager.DropApp(ctx, req)
	return err
}

func (c *rpcBasedClient) ListTables(ctx context.Context) ([]*TableInfo, error) {
	resp, err := c.metaManager.ListApps(ctx, &admin.ListAppsRequest{
		Status: admin.AppStatus_AS_AVAILABLE,
	})
	if err != nil {
		return nil, err
	}

	var results []*TableInfo
	for _, app := range resp.Infos {
		results = append(results, &TableInfo{Name: app.AppName, Envs: app.Envs})
	}
	return results, nil
}
