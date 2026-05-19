package task

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestStorage_Add(t *testing.T) {

	tests := []struct {
		Name     string
		Expected Task
		Value    AddTaskInput
	}{{
		Value:    AddTaskInput{Name: "t1", Duration: 1},
		Name:     "Add first task",
		Expected: Task{ID: 1, Name: "t1", DurationInSeconds: 1, Status: Queued},
	}, {
		Value:    AddTaskInput{Name: "t2", Duration: 3},
		Name:     "Add first task",
		Expected: Task{ID: 2, Name: "t2", DurationInSeconds: 3, Status: Queued},
	}}
	storage := NewInMemoryTaskStorage()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			before := time.Now()
			result, err := storage.Add(test.Value)
			after := time.Now()

			if err != nil {
				t.Fatal(err)
			}

			if result.ID != test.Expected.ID {
				t.Fatal("Invalid ID, got:", result.ID, "expected:", test.Expected.ID)
			}

			if result.Name != test.Expected.Name {
				t.Fatal("invalid Name, got:", result.Name, "expected:", test.Expected.Name)
			}

			if result.DurationInSeconds != test.Expected.DurationInSeconds {
				t.Fatal("Invalid Duration got:", result.DurationInSeconds, "expected:", test.Expected.DurationInSeconds)
			}

			if result.Status != test.Expected.Status {
				t.Fatal("invalid status got:", result.Status, "expected:", test.Expected.Status)
			}

			if result.CreatedAt.Before(before) || result.CreatedAt.After(after) {
				t.Fatal("invalid createdAt got:", result.CreatedAt, "expected between:", before, "and", after)
			}
		})
	}

	nameValidationTests := []struct {
		Value string
	}{
		{
			Value: "",
		},
		{
			Value: "  ",
		},
	}

	for _, test := range nameValidationTests {
		t.Run("Name Validation", func(t *testing.T) {
			storage := NewInMemoryTaskStorage()

			_, err := storage.Add(AddTaskInput{Name: test.Value, Duration: 1})

			if !errors.Is(err, ErrInvalidTaskName) {
				t.Fatal("Invalid err got:", err, "expected:", ErrInvalidTaskName)
			}
		})
	}

	t.Run("DurationInSecond validation", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		_, err := storage.Add(AddTaskInput{Name: "test"})
		if !errors.Is(err, ErrInvalidTaskDuration) {
			t.Fatal("Invalid err got:", err, "expected:", ErrInvalidTaskDuration)
		}
	})
}

func TestStorage_ConcurrentAdd(t *testing.T) {
	t.Run("Concurrent Add", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		goroutinesCount := 100
		wg := &sync.WaitGroup{}
		wg.Add(goroutinesCount)

		errChan := make(chan error, goroutinesCount)
		for i := 0; i < goroutinesCount; i++ {
			go func(i int) {
				defer wg.Done()
				_, err := storage.Add(AddTaskInput{Name: strconv.Itoa(i) + "_name", Duration: 1})
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for e := range errChan {
			if e != nil {
				t.Fatal(e)
			}
		}

		result, err := storage.GetList(GetListFilterParams{})

		if err != nil {
			t.Fatal(err)
		}

		if len(result) != goroutinesCount {
			t.Fatal("invalid count in result, got:", len(result), "expected:", goroutinesCount)
		}

		idMap := make(map[int]struct{})
		for _, item := range result {
			if item.ID <= 0 {
				t.Fatal("found zero-value id")
			}

			if _, ok := idMap[item.ID]; ok {
				t.Fatal("duplicate ID:", item.ID)
			}

			idMap[item.ID] = struct{}{}

			if item.Status != Queued {
				t.Fatal("invalid status got:", item.Status, "expected:", Queued)
			}
		}

		stats := storage.GetStats()

		if stats.Queued != goroutinesCount {
			t.Fatal("invalid queued count got:", stats.Queued, "expected:", goroutinesCount)
		}

		if stats.Done != 0 || stats.Failed != 0 || stats.Running != 0 || stats.Cancelled != 0 {
			t.Fatal("invalid stats got:", stats)
		}
	})

}

func TestStorage_GetByID(t *testing.T) {

	tests := []struct {
		Name     string
		Expected Task
		Value    AddTaskInput
	}{{
		Value:    AddTaskInput{Name: "t1", Duration: 1},
		Name:     "Add first task",
		Expected: Task{ID: 1, Name: "t1", DurationInSeconds: 1, Status: Queued},
	}}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			storage := NewInMemoryTaskStorage()

			el, err := storage.Add(test.Value)
			if err != nil {
				t.Fatal(err)
			}

			result, err := storage.GetByID(el.ID)
			if err != nil {
				t.Fatal(err)
			}

			if result.ID != test.Expected.ID {
				t.Fatal("Invalid ID, got:", result.ID, "expected:", test.Expected.ID)
			}

			if result.Name != test.Expected.Name {
				t.Fatal("invalid Name, got:", result.Name, "expected:", test.Expected.Name)
			}

			if result.DurationInSeconds != test.Expected.DurationInSeconds {
				t.Fatal("Invalid Duration got:", result.DurationInSeconds, "expected:", test.Expected.DurationInSeconds)
			}

			if result.Status != test.Expected.Status {
				t.Fatal("invalid status got:", result.Status, "expected:", test.Expected.Status)
			}
		})
	}

	t.Run("invalid ID", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()

		_, err := storage.GetByID(123)

		if !errors.Is(err, ErrTaskNotFound) {
			t.Fatal("invalid error got:", err, "expected:", ErrTaskNotFound)
		}
	})
}

func TestStorage_GetList(t *testing.T) {
	storage := NewInMemoryTaskStorage()

	_, err := storage.Add(AddTaskInput{Name: "t1", Duration: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err2 := storage.Add(AddTaskInput{Name: "t2", Duration: 1})
	if err2 != nil {
		t.Fatal(err2)
	}
	_, err3 := storage.Add(AddTaskInput{Name: "t3", Duration: 1})
	if err3 != nil {
		t.Fatal(err3)
	}
	_, err4 := storage.Add(AddTaskInput{Name: "t4", Duration: 1})
	if err4 != nil {
		t.Fatal(err4)
	}
	_, err5 := storage.Add(AddTaskInput{Name: "t5", Duration: 1})
	if err5 != nil {
		t.Fatal(err5)
	}

	_, err6 := storage.UpdateTask(Result{ID: 1, Status: Running})
	if err6 != nil {
		t.Fatal(err6)
	}
	_, err7 := storage.UpdateTask(Result{ID: 2, Status: Done})
	if err7 != nil {
		t.Fatal(err7)
	}

	_, err8 := storage.UpdateTask(Result{ID: 3, Status: Failed})
	if err8 != nil {
		t.Fatal(err8)
	}

	tests := []struct {
		Status        Status
		ExpectedCount int
	}{
		{
			Status:        Queued,
			ExpectedCount: 2,
		},
		{
			Status:        Running,
			ExpectedCount: 1,
		},
		{
			Status:        Done,
			ExpectedCount: 1,
		},
		{
			Status:        Failed,
			ExpectedCount: 1,
		},
	}

	for _, test := range tests {
		t.Run("Get list filtered by statuses ", func(t *testing.T) {
			result, err := storage.GetList(GetListFilterParams{Statuses: []Status{test.Status}})

			if err != nil {
				t.Fatal(err)
			}

			if len(result) != test.ExpectedCount {
				t.Fatal("invalid length got:", len(result), "expected:", test.ExpectedCount)
			}

			for _, item := range result {
				if item.Status != test.Status {
					t.Fatal("invalid status, got:", item.Status, "expected:", test.Status)
				}
			}
		})
	}

	testsByID := []struct {
		ID            int
		ExpectedCount int
	}{
		{
			ID:            1,
			ExpectedCount: 1,
		},
		{
			ID:            2,
			ExpectedCount: 1,
		},
		{
			ID:            3,
			ExpectedCount: 1,
		},
		{
			ID:            4,
			ExpectedCount: 1,
		},
	}

	for _, test := range testsByID {
		t.Run("Get list filtered by ID ", func(t *testing.T) {
			result, err := storage.GetList(GetListFilterParams{IDs: []int{test.ID}})

			if err != nil {
				t.Fatal(err)
			}

			if len(result) != test.ExpectedCount {
				t.Fatal("invalid length got:", len(result), "expected:", test.ExpectedCount)
			}

			for _, item := range result {
				if item.ID != test.ID {
					t.Fatal("invalid status, got:", item.ID, "expected:", test.ID)
				}
			}
		})
	}

	t.Run("Test using ID + Status", func(t *testing.T) {
		result, err := storage.GetList(GetListFilterParams{IDs: []int{1}, Statuses: []Status{Running}})
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 1 {
			t.Fatal("invalid length got:", len(result), "expected:", 1)
		}

		if result[0].ID != 1 {
			t.Fatal("invalid status, got:", result[0].ID, "expected:", 1)
		}

		if result[0].Status != Running {
			t.Fatal("invalid status, got:", result[0].Status, "expected:", Running)
		}
	})

	t.Run("Test using ID + invalid Status", func(t *testing.T) {
		result, err := storage.GetList(GetListFilterParams{IDs: []int{1}, Statuses: []Status{Queued}})
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 0 {
			t.Fatal("invalid length got:", len(result), "expected:", 0)
		}
	})

	t.Run("Test using invalid ID + Status", func(t *testing.T) {
		result, err := storage.GetList(GetListFilterParams{IDs: []int{23231}, Statuses: []Status{Running}})
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 0 {
			t.Fatal("invalid length got:", len(result), "expected:", 0)
		}
	})

	t.Run("Test using invalid ID + invalid Status", func(t *testing.T) {
		result, err := storage.GetList(GetListFilterParams{IDs: []int{23231}, Statuses: []Status{Done}})
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 0 {
			t.Fatal("invalid length got:", len(result), "expected:", 0)
		}
	})
}

func TestStorage_GetStats(t *testing.T) {
	storage := NewInMemoryTaskStorage()

	_, err := storage.Add(AddTaskInput{Name: "t1", Duration: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err2 := storage.Add(AddTaskInput{Name: "t2", Duration: 1})
	if err2 != nil {
		t.Fatal(err2)
	}
	_, err3 := storage.Add(AddTaskInput{Name: "t3", Duration: 1})
	if err3 != nil {
		t.Fatal(err3)
	}
	_, err4 := storage.Add(AddTaskInput{Name: "t4", Duration: 1})
	if err4 != nil {
		t.Fatal(err4)
	}
	_, err5 := storage.Add(AddTaskInput{Name: "t5", Duration: 1})
	if err5 != nil {
		t.Fatal(err5)
	}

	_, err6 := storage.UpdateTask(Result{ID: 1, Status: Running})
	if err6 != nil {
		t.Fatal(err6)
	}
	_, err7 := storage.UpdateTask(Result{ID: 2, Status: Done})
	if err7 != nil {
		t.Fatal(err7)
	}

	_, err8 := storage.UpdateTask(Result{ID: 3, Status: Failed})
	if err8 != nil {
		t.Fatal(err8)
	}

	t.Run("Test stats", func(t *testing.T) {
		stats := storage.GetStats()

		expected := TaskStats{
			Queued:    2,
			Running:   1,
			Done:      1,
			Failed:    1,
			Cancelled: 0,
		}

		if stats.Queued != expected.Queued || stats.Running != expected.Running || stats.Done != expected.Done || stats.Failed != expected.Failed || stats.Cancelled != expected.Cancelled {
			t.Fatal("Incalid stats, got", stats, "expected:", expected)
		}
	})
}

func TestStorage_UpdateTusk(t *testing.T) {
	tests := []struct {
		Value Status
	}{
		{
			Value: Running,
		},
		{
			Value: Done,
		},
		{
			Value: Failed,
		},
		{
			Value: Queued,
		},
		{
			Value: Cancelled,
		},
	}

	for _, test := range tests {
		t.Run(string(test.Value), func(t *testing.T) {
			storage := NewInMemoryTaskStorage()
			_, err := storage.Add(AddTaskInput{Name: "t1", Duration: 1})

			if err != nil {
				t.Fatal(err)
			}

			_, err1 := storage.UpdateTask(Result{ID: 1, Status: test.Value})
			if err1 != nil {
				t.Fatal(err1)
			}

			task, err2 := storage.GetByID(1)
			if err2 != nil {
				t.Fatal(err2)
			}

			if task.Status != test.Value {
				t.Fatal("invalid status got:", task.Status, "expected:", test.Value)
			}
		})
	}

	t.Run("invalid Status", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		_, err := storage.Add(AddTaskInput{Name: "t1", Duration: 1})
		if err != nil {
			t.Fatal(err)
		}

		_, err1 := storage.UpdateTask(Result{ID: 1, Status: "sdsd123"})

		if !errors.Is(err1, ErrInvalidStatus) {
			t.Fatal("invalid Error got:", err1, "expected:", ErrInvalidStatus)
		}
	})

	t.Run("invalid Id", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		_, err := storage.Add(AddTaskInput{Name: "t1", Duration: 1})
		if err != nil {
			t.Fatal(err)
		}
		_, err1 := storage.UpdateTask(Result{ID: 123, Status: Queued})

		if !errors.Is(err1, ErrTaskNotFound) {
			t.Fatal("invalid Error got:", err1, "expected:", ErrTaskNotFound)
		}
	})
}

func TestStorage_ConcurrentUpdateAndRead(t *testing.T) {
	storage := NewInMemoryTaskStorage()
	goroutinesCount := 100
	goroutinesReadCount := 50
	errChan := make(chan error, goroutinesCount+goroutinesReadCount)

	var tasks []Task

	for i := 0; i < goroutinesCount; i++ {
		task, err := storage.Add(AddTaskInput{
			Name:     strconv.Itoa(i) + "_task",
			Duration: 1,
		})
		if err != nil {
			t.Fatal(err)
		}

		tasks = append(tasks, task)
	}

	wg := sync.WaitGroup{}
	wg.Add(goroutinesCount + goroutinesReadCount)

	for _, task := range tasks {
		go func(task Task) {
			defer wg.Done()
			_, err := storage.UpdateTask(Result{
				ID:     task.ID,
				Status: Done,
			})

			if err != nil {
				errChan <- err
			}
		}(task)
	}

	for i := 0; i < goroutinesReadCount; i++ {
		go func(i int) {
			defer wg.Done()
			storage.GetStats()
			_, err := storage.GetList(GetListFilterParams{})
			if err != nil {
				errChan <- err
			}
			_, err2 := storage.GetByID(tasks[i].ID)
			if err2 != nil {
				errChan <- err2
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Fatal(err)
	}

	stats := storage.GetStats()

	if stats.Queued != 0 || stats.Done != goroutinesCount || stats.Failed != 0 || stats.Cancelled != 0 || stats.Running != 0 {
		t.Fatal("unexpected stats:", stats)
	}
}
