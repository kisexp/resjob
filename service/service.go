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
