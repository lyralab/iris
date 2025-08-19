package scheduler

type ServiceInterface interface {
	// Start initializes the scheduler service.
	Start() error
	// Stop gracefully stops the scheduler service.
	Stop() error
}
