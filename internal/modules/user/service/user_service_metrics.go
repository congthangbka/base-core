package service

import (
	"context"
	"time"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/metrics"
	"llm-aggregator/internal/modules/user/dto"
)

// instrumentedUserService wraps UserService with metrics
type instrumentedUserService struct {
	service UserService
}

// NewInstrumentedUserService creates a new instrumented service wrapper
func NewInstrumentedUserService(service UserService) UserService {
	return &instrumentedUserService{
		service: service,
	}
}

func (s *instrumentedUserService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	start := time.Now()
	result, err := s.service.Create(ctx, req)
	duration := time.Since(start).Seconds()

	metrics.BusinessOperationsTotal.WithLabelValues("create", "user").Inc()
	metrics.BusinessOperationDuration.WithLabelValues("create", "user").Observe(duration)

	if err != nil {
		errorCode := "unknown"
		if svcErr, ok := err.(*common.ServiceError); ok {
			errorCode = svcErr.Code
		}
		metrics.BusinessErrorsTotal.WithLabelValues("create", "user", errorCode).Inc()
	}

	return result, err
}

func (s *instrumentedUserService) Update(ctx context.Context, id string, req *dto.UpdateUserRequest) error {
	start := time.Now()
	err := s.service.Update(ctx, id, req)
	duration := time.Since(start).Seconds()

	metrics.BusinessOperationsTotal.WithLabelValues("update", "user").Inc()
	metrics.BusinessOperationDuration.WithLabelValues("update", "user").Observe(duration)

	if err != nil {
		errorCode := "unknown"
		if svcErr, ok := err.(*common.ServiceError); ok {
			errorCode = svcErr.Code
		}
		metrics.BusinessErrorsTotal.WithLabelValues("update", "user", errorCode).Inc()
	}

	return err
}

func (s *instrumentedUserService) Delete(ctx context.Context, id string) error {
	start := time.Now()
	err := s.service.Delete(ctx, id)
	duration := time.Since(start).Seconds()

	metrics.BusinessOperationsTotal.WithLabelValues("delete", "user").Inc()
	metrics.BusinessOperationDuration.WithLabelValues("delete", "user").Observe(duration)

	if err != nil {
		errorCode := "unknown"
		if svcErr, ok := err.(*common.ServiceError); ok {
			errorCode = svcErr.Code
		}
		metrics.BusinessErrorsTotal.WithLabelValues("delete", "user", errorCode).Inc()
	}

	return err
}

func (s *instrumentedUserService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	start := time.Now()
	result, err := s.service.GetByID(ctx, id)
	duration := time.Since(start).Seconds()

	metrics.BusinessOperationsTotal.WithLabelValues("get_by_id", "user").Inc()
	metrics.BusinessOperationDuration.WithLabelValues("get_by_id", "user").Observe(duration)

	if err != nil {
		errorCode := "unknown"
		if svcErr, ok := err.(*common.ServiceError); ok {
			errorCode = svcErr.Code
		}
		metrics.BusinessErrorsTotal.WithLabelValues("get_by_id", "user", errorCode).Inc()
	}

	return result, err
}

func (s *instrumentedUserService) GetAll(ctx context.Context, req *dto.PagingRequest) (*dto.UserPagingResponse, error) {
	start := time.Now()
	result, err := s.service.GetAll(ctx, req)
	duration := time.Since(start).Seconds()

	metrics.BusinessOperationsTotal.WithLabelValues("get_all", "user").Inc()
	metrics.BusinessOperationDuration.WithLabelValues("get_all", "user").Observe(duration)

	if err != nil {
		errorCode := "unknown"
		if svcErr, ok := err.(*common.ServiceError); ok {
			errorCode = svcErr.Code
		}
		metrics.BusinessErrorsTotal.WithLabelValues("get_all", "user", errorCode).Inc()
	}

	return result, err
}
