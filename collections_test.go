package main

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Error("Bad constructor")
	}
}

func TestQueueEnqueue(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		q.QueueEnqueue(i)
	}

	if q.QueueSize() != 10 {
		t.Error("expected size 10, got", q.QueueSize())
	}
}

func TestQueueDequeue(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		q.QueueEnqueue(i)
	}

	for i := 10; i > 0; i-- {
		val, err := q.QueueDequeue()
		if err != nil {
			t.Error(err)
		}
		if (*val).(int) != (10 - i) {
			t.Error("expected", 10-i, "got ", val)
		}

		if q.QueueSize() != int64(i-1) {
			t.Error("expected size", i-1, "got", q.QueueSize())
		}
	}
}
