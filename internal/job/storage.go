package job

import "github.com/google/uuid"

//go:generate mockery --case underscore --name Storage
type Storage interface {
	Store(job *Job) error
	GetByName(name string) (*Job, error)
	GetAll() ([]Job, error)
	DeleteByName(name string) error
}

//go:generate mockery --case underscore --name ExecutionStorage
type ExecutionStorage interface {
	Store(execution *Execution) error
	GetByJobName(jobName string) ([]Execution, error)
	GetByID(id uuid.UUID) (*Execution, error)
	DeleteByJobName(jobName string) error
	Delete(executionID uuid.UUID) error
}
