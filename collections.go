package main

import (
	"errors"
)

type (
	node struct {
		value interface{}
		next  *node
	}

	//Queue is...
	Queue struct {
		first, last *node
		size        int64
	}
)

//NewQueue makes a new queue
func NewQueue() *Queue {
	return &Queue{}
}

//QueueEnqueue adds an element which is non-nil to the tail of the queue
func (q *Queue) QueueEnqueue(value interface{}) (int64, error) {
	if value == nil {
		return q.size, errors.New("queue value can't be null")
	}
	n := &node{value, nil}
	if q.size == 0 {
		q.first = n
		q.last = n
	} else {
		q.first.next = n
		q.last = n
	}
	q.size++

	return q.size, nil
}

//QueueDequeue pulls the first element out of the queue, if there is one
func (q *Queue) QueueDequeue() (*interface{}, error) {
	if q.size == 0 {
		return nil, errors.New("queue has no elements")
	}

	resVal := &q.first.value

	q.first = q.first.next
	q.size--

	if q.size == 0 {
		q.first = nil
		q.last = nil
	}

	return resVal, nil
}

//QueueSize returns the size of the queue
func (q *Queue) QueueSize() int64 {
	return q.size
}
