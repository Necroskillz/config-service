package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) Authenticate(ctx context.Context, name, password string) (*model.User, error) {
	user, err := s.userRepository.GetByProperty(ctx, "name", name)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, nil
	}

	return user, nil
}

func (s *UserService) Get(ctx context.Context, id uint) (*model.User, error) {
	return s.userRepository.GetById(ctx, id, "Permissions", "Permissions.VariationPropertyValues")
}
