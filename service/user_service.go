package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      repository.Repository
	jwtSecret string
}

func NewUserService(repo repository.Repository, config *utils.Config) *UserService {
	jwtSecret := config.JWTSecret
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var jurisdiction *string
	var barNumber *string
	if req.Jurisdiction != "" {
		jurisdiction = &req.Jurisdiction
	}
	if req.BarNumber != "" {
		barNumber = &req.BarNumber
	}

	user, err := s.repo.CreateUser(ctx, &repository.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         utils.ToPgtypeText(&req.Name),
		Role:         req.Role,
		Jurisdiction: utils.ToPgtypeText(jurisdiction),
		BarNumber:    utils.ToPgtypeText(barNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			Name:         utils.GetStringOrEmpty(utils.GetNullableString(user.Name)),
			Role:         user.Role,
			Jurisdiction: jurisdiction,
			BarNumber:    barNumber,
			CreatedAt:    utils.PgtypeTimeToTime(user.CreatedAt),
		},
	}, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var jurisdiction *string
	var barNumber *string
	if user.Jurisdiction.Valid {
		jurisdiction = &user.Jurisdiction.String
	}
	if user.BarNumber.Valid {
		barNumber = &user.BarNumber.String
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			Name:         utils.GetStringOrEmpty(utils.GetNullableString(user.Name)),
			Role:         user.Role,
			Jurisdiction: jurisdiction,
			BarNumber:    barNumber,
			CreatedAt:    utils.PgtypeTimeToTime(user.CreatedAt),
		},
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var jurisdiction *string
	var barNumber *string
	if user.Jurisdiction.Valid {
		jurisdiction = &user.Jurisdiction.String
	}
	if user.BarNumber.Valid {
		barNumber = &user.BarNumber.String
	}

	return &dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         utils.GetStringOrEmpty(utils.GetNullableString(user.Name)),
		Role:         user.Role,
		Jurisdiction: jurisdiction,
		BarNumber:    barNumber,
		CreatedAt:    utils.PgtypeTimeToTime(user.CreatedAt),
	}, nil
}
