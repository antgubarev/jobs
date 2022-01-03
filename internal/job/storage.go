package job

import "github.com/google/uuid"

//go:generate mockery --case underscore --inpackage --name JobStorage
type JobStorage interface {
	Store(job *Job) error
	GetByName(name string) (*Job, error)
	GetAll() ([]Job, error)
	DeleteByName(name string) error
}

//go:generate mockery --case underscore --inpackage --name ExecutionStorage
type ExecutionStorage interface {
	Store(execution *Execution) error
	GetByJobName(jobName string) ([]Execution, error)
	GetById(id uuid.UUID) (Execution, error)
	DeleteByJobName(jobName string) error
	Delete(execution *Execution) error
}
