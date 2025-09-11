package orders

type Service struct {
	psqlRepo  Repository
	cacheRepo Repository
}

// Конструктор для создания нового Service
func NewService(psqlRepo Repository, cacheRepo Repository) *Service {
	return &Service{
		psqlRepo:  psqlRepo,
		cacheRepo: cacheRepo,
	}
}

// Сохраняет заказ в оба репозитория
func (s *Service) SaveOrder(order *ModelOrder) error {
	if err := s.cacheRepo.Save(order); err != nil {
		return err
	}
	if err := s.psqlRepo.Save(order); err != nil {
		return err
	}
	return nil
}

// Получает заказ по ID, сначала ищет в кэше, если не находит — в БД
func (s *Service) GetOrderByID(id string) (*ModelOrder, error) {
	order, err := s.cacheRepo.GetByID(id)
	if err == nil {
		return order, nil
	}
	return s.psqlRepo.GetByID(id)
}

// Получает все заказы из БД
func (s *Service) GetAllOrders() ([]*ModelOrder, error) {
	return s.psqlRepo.GetAll()
}
