package service

import (
	"context"

	"github.com/gogazub/myapp/internal/model"
	repo "github.com/gogazub/myapp/internal/repository"
)

type IService interface {
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
}

type Service struct {
	psqlRepo  repo.IDBRepository
	cacheRepo repo.ICacheRepository
}

// Конструктор для создания нового Service
func NewService(psqlRepo repo.IDBRepository, cacheRepo repo.ICacheRepository) *Service {
	return &Service{
		psqlRepo:  psqlRepo,
		cacheRepo: cacheRepo,
	}
}

// Сохраняет заказ в оба репозитория
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	if err := s.psqlRepo.Save(ctx, order); err != nil {
		return err
	}
	if err := s.cacheRepo.Save(ctx, order); err != nil {
		return err
	}
	return nil
}

// Получает заказ по ID, сначала ищет в кэше, если не находит — в БД
func (s *Service) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	order, err := s.cacheRepo.GetByID(ctx, id)
	if err == nil {
		return order, nil
	}
	order, err = s.psqlRepo.GetByID(ctx, id)
	if err == nil {
		if order != nil {
			s.cacheRepo.Save(ctx, order)
		}
	}
	return order, err
}
