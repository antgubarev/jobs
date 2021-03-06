// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	job "github.com/antgubarev/jobs/internal/job"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// ExecutionStorage is an autogenerated mock type for the ExecutionStorage type
type ExecutionStorage struct {
	mock.Mock
}

// Delete provides a mock function with given fields: executionID
func (_m *ExecutionStorage) Delete(executionID uuid.UUID) error {
	ret := _m.Called(executionID)

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) error); ok {
		r0 = rf(executionID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByJobName provides a mock function with given fields: jobName
func (_m *ExecutionStorage) DeleteByJobName(jobName string) error {
	ret := _m.Called(jobName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(jobName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: id
func (_m *ExecutionStorage) GetByID(id uuid.UUID) (*job.Execution, error) {
	ret := _m.Called(id)

	var r0 *job.Execution
	if rf, ok := ret.Get(0).(func(uuid.UUID) *job.Execution); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.Execution)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uuid.UUID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByJobName provides a mock function with given fields: jobName
func (_m *ExecutionStorage) GetByJobName(jobName string) ([]job.Execution, error) {
	ret := _m.Called(jobName)

	var r0 []job.Execution
	if rf, ok := ret.Get(0).(func(string) []job.Execution); ok {
		r0 = rf(jobName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]job.Execution)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(jobName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: execution
func (_m *ExecutionStorage) Store(execution *job.Execution) error {
	ret := _m.Called(execution)

	var r0 error
	if rf, ok := ret.Get(0).(func(*job.Execution) error); ok {
		r0 = rf(execution)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
