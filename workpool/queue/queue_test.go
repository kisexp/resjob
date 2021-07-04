package queue

import (
	"fmt"
	"testing"
)

func TestQueueSimple(t *testing.T) {
	q := New()
	for i := 0; i < minQueueLen; i++ {
		fmt.Println("i -> ", i)
		q.Add(i)
	}

	for i := 0; i < minQueueLen; i++ {
		fmt.Println("peek ->", q.Peek())
		if q.Peek().(int) != i {
			t.Error("peek", i, "had value", q.Peek())
		}
		x := q.Remove()
		fmt.Println("remove -> ", x)
		if x != i {
			t.Error("remove", i, "had value", x)
		}
	}
}

func TestQueueWrapping(t *testing.T) {
	q := New()
	for i := 0; i < minQueueLen; i++ {
		q.Add(i)
	}
	for i := 0; i < 3; i++ {
		q.Remove()
		q.Add(minQueueLen + i)
		fmt.Println("add -> ", minQueueLen+i)
	}

	for i := 0; i < minQueueLen; i++ {
		fmt.Println("peek -> ", q.Peek())
		if q.Peek().(int) != i+3 {
			t.Error("peek", i, "had value", q.Peek())
		}
		q.Remove()
		fmt.Println("remove -> ", i)
	}
}

func TestQueueLength(t *testing.T) {
	q := New()
	fmt.Println("length -> ", q.Length())
	if q.Length() != 0 {
		t.Error("empty queue length not 0")
	}

	for i := 0; i < 1000; i++ {
		q.Add(i)
		fmt.Println("for length -> ", q.Length(), i+1)
		if q.Length() != i+1 {
			t.Error("adding: queue with", i, "elements has length", q.Length())
		}

	}
	for i := 0; i < 1000; i++ {
		q.Remove()
		if q.Length() != 1000-i-1 {
			t.Error("removing: queue with", 1000-i-i, "elements has length", q.Length())
		}
	}
}

func TestQueueGet(t *testing.T) {
	q := New()
	for i := 0; i < 1000; i++ {
		q.Add(i)
		for j := 0; j < q.Length(); j++ {
			if q.Get(j).(int) != j {
				t.Errorf("index %d doesn't contain %d", j, j)
			}
		}
	}

}

func TestQueueGetNegative(t *testing.T) {
	q := New()
	for i := 0; i < 1000; i++ {
		q.Add(i)
		for j := 1; j <= q.Length(); j++ {
			if q.Get(-j).(int) != q.Length()-j {
				t.Errorf("index %d donesn't contain %d", -j, q.Length()-j)
			}
		}
	}
}
