// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	job "github.com/antgubarev/jobs/internal/job"
	mock "github.com/stretchr/testify/mock"
)

// JobStorage is an autogenerated mock type for the JobStorage type
type JobStorage struct {
	mock.Mock
}

// DeleteByName provides a mock function with given fields: name
func (_m *JobStorage) DeleteByName(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields:
func (_m *JobStorage) GetAll() ([]job.Job, error) {
	ret := _m.Called()

	var r0 []job.Job
	if rf, ok := ret.Get(0).(func() []job.Job); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]job.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByName provides a mock function with given fields: name
func (_m *JobStorage) GetByName(name string) (*job.Job, error) {
	ret := _m.Called(name)

	var r0 *job.Job
	if rf, ok := ret.Get(0).(func(string) *job.Job); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: _a0
func (_m *JobStorage) Store(_a0 *job.Job) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*job.Job) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
