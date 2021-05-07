// Package service 使用说明：
// 1.初始化dao中各个结构体
// 2.处理具体的业务逻辑
// 3.组装dao中各种封装好的方法，如先调用缓存获取后读取数据库等
// 4.是整个服务的枢纽，用于初始化、组装、调度、管理各种资源（缓存、数据库、grpc链接等）
package service

import (
	"twist/config"
	"twist/dao"
)

// Service .
type Service struct {
	Config *config.Config
	Dao    *dao.Dao
}

// NewService NewService
func NewService(c *config.Config) (*Service, error) {
	d, err := dao.New(c)
	if err != nil {
		return nil, err
	}

	svr := &Service{
		Config: c,
		Dao:    d,
	}

	return svr, nil
}

// Close all
func (s *Service) Close() {
	s.Dao.Close()
}
