package server

import (
	"errors"
	"fmt"
	"gateway/internal/conf"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdRegister(conf *conf.Registry) (registry.Registrar, error) {
	if conf.Etcd == nil {
		return nil, errors.New("etcd 配置为空")
	}
	timeout, err := time.ParseDuration(conf.Etcd.Timeout)
	if err != nil {
		panic(fmt.Errorf("invalid etcd timeout: %v", err))
	}
	client, err1 := clientv3.New(
		clientv3.Config{
			Endpoints:   conf.Etcd.Endpoints,
			DialTimeout: timeout,
		},
	)
	if err1 != nil {
		log.Errorf("连接 etcd 失败: %v，地址: %v", err1, conf.Etcd.Endpoints)
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
