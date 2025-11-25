package server

import (
	"errors"
	"fmt"
	"seckill_service/internal/conf"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdRegister(conf *conf.Registry) (registry.Registrar, error) {
	timeout, err := time.ParseDuration(conf.Etcd.Timeout)
	if err != nil {
		return nil, err
	}
	client, err1 := clientv3.New(clientv3.Config{Endpoints: conf.Etcd.Endpoints, DialTimeout: timeout})
	if err1 != nil {
		return nil, err1
	}

	r := etcd.New(client)
	if r == nil {
		log.Errorf("创建 etcd 注册器失败")
		return nil, errors.New("创建 etcd 注册器失败")
	}

	return r, nil
}

func NewEtcdDiscovery(conf *conf.Registry) (registry.Discovery, error) {
	timeout, err := time.ParseDuration(conf.Etcd.Timeout)
	if err != nil {
		panic(fmt.Errorf("invalid etcd timeout: %v", err))
	}
	cli, err := clientv3.New(
		clientv3.Config{
			Endpoints:   conf.Etcd.Endpoints,
			DialTimeout: timeout,
		},
	)
	if err != nil {
		return nil, err
	}
	r := etcd.New(cli)
	return r, nil
}
