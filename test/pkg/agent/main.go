// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	ctntcfg "github.com/bhojpur/platform/content-service/api/config"
	"github.com/bhojpur/platform/content-service/pkg/storage"
	"github.com/bhojpur/platform/test/pkg/agent/daemon/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func main() {
	integration.ServeAgent(new(DaemonAgent))
}

type daemonConfig struct {
	Daemon struct {
		Content struct {
			Storage ctntcfg.StorageConfig `json:"storage"`
		} `json:"content"`
	} `json:"daemon"`
}

// DaemonAgent provides ingteration test services from within bp-daemon
type DaemonAgent struct {
}

// CreateBucket reads the daemon's config, and creates a bucket
func (*DaemonAgent) CreateBucket(args *api.CreateBucketRequest, resp *api.CreateBucketResponse) error {
	*resp = api.CreateBucketResponse{}

	fc, err := os.ReadFile("/config/config.json")
	if err != nil {
		return err
	}
	var cfg daemonConfig
	err = json.Unmarshal(fc, &cfg)
	if err != nil {
		return err
	}

	ac, err := storage.NewDirectAccess(&cfg.Daemon.Content.Storage)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err = ac.Init(ctx, args.Owner, args.Application, "")
	if err != nil {
		return err
	}

	err = ac.EnsureExists(ctx)
	if err != nil {
		return err
	}

	return nil
}
