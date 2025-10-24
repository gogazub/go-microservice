// Package service реализует сервис, инкапсулирующий бизнес логику
package service

import (
	"context"
	"log"

	"github.com/gogazub/myapp/internal/model"
	repo "github.com/gogazub/myapp/internal/repository"
)

// IService интерфейс сервиса
type IService interface {
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
}

// Service реализация сервиса.
type Service struct {
	psqlRepo  repo.IDBRepository
	cacheRepo repo.ICacheRepository
}

// NewService конструктор нового Service
func NewService(psqlRepo repo.IDBRepository, cacheRepo repo.ICacheRepository) *Service {
	return &Service{
		psqlRepo:  psqlRepo,
		cacheRepo: cacheRepo,
	}
}

// SaveOrder Сохраняет заказ в кеш и в БД
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	if err := s.psqlRepo.Save(ctx, order); err != nil {
		return err
	}
	if err := s.cacheRepo.Save(ctx, order); err != nil {
		return err
	}
	return nil
}

// GetOrderByID Cache-Aside поиск заказа по id
func (s *Service) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	order, err := s.cacheRepo.GetByID(ctx, id)
	if err == nil {
		return order, nil
	}
	order, err = s.psqlRepo.GetByID(ctx, id)
	if err == nil {
		if order != nil {
			err := s.cacheRepo.Save(ctx, order)
			if err != nil {
				log.Printf("cacheRepo save order error:%s", err.Error())
			}
		}
	}
	return order, err
}
