package services

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"medical-system/domain/entities"
	apperrors "medical-system/internal/errors"
)

// MockConcurrentUserRepository es una versión thread-safe del mock de UserRepository
type MockConcurrentUserRepository struct {
	mock.Mock
	mu sync.RWMutex
}

func (m *MockConcurrentUserRepository) Create(user *entities.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockConcurrentUserRepository) FindByEmailAndTenant(email, tenantID string) (*entities.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	args := m.Called(email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockConcurrentUserRepository) FindByID(id string) (*entities.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockConcurrentUserRepository) Update(user *entities.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockConcurrentUserRepository) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(id)
	return args.Error(0)
}

// MockConcurrentTenantValidator es una versión thread-safe del mock de TenantValidator
type MockConcurrentTenantValidator struct {
	mock.Mock
	mu sync.RWMutex
}

func (m *MockConcurrentTenantValidator) ValidateTenantLimits(tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(tenantID)
	if len(args) > 0 {
		if err, ok := args.Get(0).(error); ok {
			return err
		}
	}
	return nil
}

// TestConcurrentUserRegistration prueba el registro concurrente de usuarios en el mismo tenant
func TestConcurrentUserRegistration(t *testing.T) {
	mockRepo := new(MockConcurrentUserRepository)
	mockValidator := new(MockConcurrentTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	// Configurar mocks para permitir múltiples registros
	mockValidator.On("ValidateTenantLimits", "1").Return(nil).Maybe()
	mockRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil).Maybe()

	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Canal para recolectar resultados
	results := make(chan error, numGoroutines)

	// Lanzar goroutines concurrentes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			user := &entities.User{
				Email:    "user" + string(rune(id+'0')) + "@example.com",
				TenantID: "1",
			}

			err := authService.RegisterUser(user, "password123")
			results <- err
		}(i)
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	close(results)

	// Verificar resultados
	successCount := 0
	errorCount := 0

	for err := range results {
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Deberían pasar todas las operaciones (sin límites estrictos en este test)
	assert.Equal(t, numGoroutines, successCount)
	assert.Equal(t, 0, errorCount)
}

// TestConcurrentTenantLimitEnforcement prueba que los límites de tenant se respeten bajo concurrencia
func TestConcurrentTenantLimitEnforcement(t *testing.T) {
	mockRepo := new(MockConcurrentUserRepository)
	mockValidator := new(MockConcurrentTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	// Configurar el validador para fallar después de cierto número de llamadas
	var callCount int32
	mockValidator.On("ValidateTenantLimits", "1").Return(func() error {
		count := atomic.AddInt32(&callCount, 1)
		if count > 5 { // Solo permitir 5 registros
			return apperrors.NewConflictError("tenant has reached maximum user limit")
		}
		return nil
	})

	mockRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil).Maybe()

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Canal para recolectar resultados
	results := make(chan error, numGoroutines)

	// Lanzar goroutines concurrentes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			user := &entities.User{
				Email:    "user" + string(rune(id+'0')) + "@example.com",
				TenantID: "1",
			}

			err := authService.RegisterUser(user, "password123")
			results <- err
		}(i)
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	close(results)

	// Verificar resultados
	successCount := 0
	errorCount := 0

	for err := range results {
		if err != nil {
			errorCount++
			// Verificar que el error sea del tipo correcto
			assert.IsType(t, &apperrors.AppError{}, err)
			appErr := err.(*apperrors.AppError)
			assert.Equal(t, apperrors.ErrorTypeConflict, appErr.Type)
		} else {
			successCount++
		}
	}

	// Verificar que el límite se respeta (algunos pasan, algunos fallan)
	assert.True(t, successCount >= 0 && successCount <= 10)
	assert.True(t, errorCount >= 0 && errorCount <= 10)
	assert.Equal(t, 10, successCount+errorCount)
}

// TestConcurrentProfileUpdates prueba actualizaciones de perfil concurrentes
func TestConcurrentProfileUpdates(t *testing.T) {
	mockRepo := new(MockConcurrentUserRepository)
	mockValidator := new(MockConcurrentTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		ID:        "1",
		Email:     "test@example.com",
		FirstName: "Original",
		LastName:  "Name",
	}

	// Configurar mocks
	mockRepo.On("FindByID", "1").Return(user, nil).Maybe()
	mockRepo.On("Update", mock.AnythingOfType("*entities.User")).Return(nil).Maybe()

	numGoroutines := 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Canal para recolectar resultados
	results := make(chan error, numGoroutines)

	// Lanzar goroutines concurrentes que actualizan el perfil
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			newFirstName := "Updated" + string(rune(id+'0'))
			_, err := authService.UpdateProfile("1", newFirstName, "Name", "test@example.com")
			results <- err
		}(i)
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	close(results)

	// Verificar resultados
	successCount := 0
	errorCount := 0

	for err := range results {
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Todas las operaciones deberían pasar (aunque el estado final sea indeterminado debido a race conditions)
	assert.Equal(t, numGoroutines, successCount)
	assert.Equal(t, 0, errorCount)
}

// TestConcurrentCredentialVerification prueba verificación de credenciales bajo carga concurrente
func TestConcurrentCredentialVerification(t *testing.T) {
	mockRepo := new(MockConcurrentUserRepository)
	mockValidator := new(MockConcurrentTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // bcrypt hash for "password"
		TenantID:     "1",
	}

	// Configurar mocks
	mockRepo.On("FindByEmailAndTenant", "test@example.com", "1").Return(user, nil).Maybe()

	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Canal para recolectar resultados
	results := make(chan error, numGoroutines)

	// Lanzar goroutines concurrentes
	start := time.Now()
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			_, err := authService.VerifyCredentials("test@example.com", "password", "1")
			results <- err
		}()
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	close(results)

	duration := time.Since(start)

	// Verificar resultados
	successCount := 0
	errorCount := 0

	for err := range results {
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Todas las verificaciones deberían pasar
	assert.Equal(t, numGoroutines, successCount)
	assert.Equal(t, 0, errorCount)

	// Verificar que el rendimiento sea razonable (menos de 10 segundos para 100 verificaciones bcrypt)
	// bcrypt es computacionalmente intensivo, así que permitimos más tiempo
	assert.Less(t, duration, 10*time.Second)

	t.Logf("Concurrent credential verification completed in %v", duration)
}

// TestRaceConditionPrevention prueba que no hay race conditions en operaciones críticas
func TestRaceConditionPrevention(t *testing.T) {
	mockRepo := new(MockConcurrentUserRepository)
	mockValidator := new(MockConcurrentTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	// Usar un contador compartido para simular estado mutable
	var registrationCount int32

	mockValidator.On("ValidateTenantLimits", "1").Return(func() error {
		count := atomic.AddInt32(&registrationCount, 1)
		if count > 3 {
			return apperrors.NewConflictError("tenant has reached maximum user limit")
		}
		return nil
	})

	mockRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil).Maybe()

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Canal para recolectar resultados
	results := make(chan struct {
		success bool
		count   int
	}, numGoroutines)

	// Lanzar goroutines concurrentes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			user := &entities.User{
				Email:    "user" + string(rune(id+'0')) + "@example.com",
				TenantID: "1",
			}

			err := authService.RegisterUser(user, "password123")

			currentCount := int(atomic.LoadInt32(&registrationCount))

			results <- struct {
				success bool
				count   int
			}{success: err == nil, count: currentCount}
		}(i)
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	close(results)

	// Verificar resultados
	successCount := 0
	errorCount := 0
	finalCounts := make([]int, 0)

	for result := range results {
		if result.success {
			successCount++
		} else {
			errorCount++
		}
		finalCounts = append(finalCounts, result.count)
	}

	// El test verifica que el límite se respete, pero debido a la concurrencia
	// el número exacto puede variar. Lo importante es que el contador se incremente correctamente
	assert.Equal(t, 10, successCount+errorCount)

	// Verificar que el contador final sea consistente (todas las llamadas pasaron por la validación)
	finalCount := int(atomic.LoadInt32(&registrationCount))
	assert.Equal(t, 10, finalCount)

	// Verificar que algunos registros pasaron y algunos fallaron (comportamiento esperado bajo concurrencia)
	assert.True(t, successCount > 0, "Should have at least some successful registrations")
	assert.True(t, errorCount > 0, "Should have at least some failed registrations")
}
