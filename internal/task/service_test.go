package task

import (
	"errors"
	"testing"
)

func TestTaskService_Add(t *testing.T) {
	t.Run("Happy case", func(t *testing.T) {
		service := NewTaskService(NewInMemoryTaskStorage())

		item, err := service.Add(AddTaskInput{Name: "Test task", Duration: 1})

		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if item.Name != "Test task" {
			t.Fatal("Expected item to be created with name 'Test task', got:", item.Name)
		}

		if item.DurationInSeconds != 1 {
			t.Fatal("Expected item to be created with DurationInSeconds '1', got:", item.DurationInSeconds)
		}
	})
}

func TestTaskService_GetByID(t *testing.T) {
	t.Run("Happy case", func(t *testing.T) {
		service := NewTaskService(NewInMemoryTaskStorage())

		item, err := service.Add(AddTaskInput{Name: "Test task", Duration: 1})

		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		result, err1 := service.GetByID(item.ID)

		if err1 != nil {
			t.Fatal("Expected no error, got:", err1)
		}

		if result.Name != "Test task" {
			t.Fatal("Expected item to be created with name 'Test task', got:", result.Name)
		}

		if result.DurationInSeconds != 1 {
			t.Fatal("Expected item to be created with DurationInSeconds '1', got:", result.DurationInSeconds)
		}
	})

}

func TestTaskService_UpdateTask_happy1(t *testing.T) {

	service := NewTaskService(NewInMemoryTaskStorage())

	t.Run("Succesful update", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "Test task", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		result, err := service.UpdateTask(Result{ID: item.ID, Status: Running, Error: errors.New("Some err")})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if result.ID != item.ID {
			t.Fatal("Expected item ID to be the same, got:", result.ID)
		}

		if result.Status != Running {
			t.Fatal("Expected item Status to be the same, got:", result.ID)
		}
	})
}

func TestTaskService_UpdateTask_happy2(t *testing.T) {
	happyStatuses := []struct {
		Value Status
	}{
		{Value: Running},
		{Value: Done},
	}
	service := NewTaskService(NewInMemoryTaskStorage())
	happyitem, err := service.Add(AddTaskInput{Name: "1", Duration: 1})
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	for _, status := range happyStatuses {
		name := "Happy case:" + string(status.Value)
		t.Run(name, func(t *testing.T) {
			_, err := service.UpdateTask(Result{ID: happyitem.ID, Status: status.Value})
			if err != nil {
				t.Fatal("Expected no error, got:", err)
			}
		})
	}
}

func TestTaskService_UpdateTask_happy3(t *testing.T) {
	service := NewTaskService(NewInMemoryTaskStorage())

	happyitem, err := service.Add(AddTaskInput{Name: "1", Duration: 1})
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	happyStatuses := []struct {
		Value Status
	}{
		{Value: Cancelled},
	}
	for _, status := range happyStatuses {
		t.Run("Happy case", func(t *testing.T) {
			_, err := service.UpdateTask(Result{ID: happyitem.ID, Status: status.Value})
			if err != nil {
				t.Fatal("Expected no error, got:", err)
			}
		})
	}

}

func TestTaskService_UpdateTask_happy4(t *testing.T) {
	service := NewTaskService(NewInMemoryTaskStorage())

	happyitem, err := service.Add(AddTaskInput{Name: "1", Duration: 1})
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	happyStatuses := []struct {
		Value Status
	}{
		{Value: Running},
		{Value: Failed},
	}
	for _, status := range happyStatuses {
		t.Run("Happy case", func(t *testing.T) {
			_, err := service.UpdateTask(Result{ID: happyitem.ID, Status: status.Value})
			if err != nil {
				t.Fatal("Expected no error, got:", err)
			}
		})
	}
}

func TestTaskService_UpdateTask_negative(t *testing.T) {
	service := NewTaskService(NewInMemoryTaskStorage())

	t.Run("Invalid transaction not change the state", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "1", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		service.UpdateTask(Result{ID: item.ID, Status: Done})
		r, err1 := service.GetByID(item.ID)
		if err1 != nil {
			t.Fatal("Expected no error, got:", err1)
		}
		if r.Status != item.Status {
			t.Fatal("Expected ", r.Status, "got:", item.Status)
		}
	})

	t.Run("Invalid Id", func(t *testing.T) {
		_, err := service.UpdateTask(Result{ID: 12312344123, Status: Done})
		if !errors.Is(err, ErrTaskNotFound) {
			t.Fatal("Expected ", ErrTaskNotFound, "got:", err)
		}
	})

	t.Run("From Queued to Done", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "1", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Queued to Failed", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "2", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Queued to Queued", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "3", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Queued})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	//

	t.Run("From Running to Queued", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "4", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Queued})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Running to Running", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "5", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Done to Queued", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "6", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Queued})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Done to Running", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "7", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Done to Failed", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "8", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Done to Cancelled", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "9", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Done to Done", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "10", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("Failed Failed to Queued", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "11", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Queued})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Failed to Running", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "12", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Failed to Done", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "13", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Failed to Cancelled", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "14", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Failed to Failed", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "15", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err3 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if err3 != nil {
			t.Fatal("Expected no error, got:", err3)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})
	///

	t.Run("From Cancelled to Queued", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "16", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Queued})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Cancelled to Running", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "17", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Running})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Cancelled to Done", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "18", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Done})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Cancelled to Failed", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "19", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Failed})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})

	t.Run("From Cancelled to Canceled", func(t *testing.T) {
		item, err := service.Add(AddTaskInput{Name: "20", Duration: 1})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		_, err2 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if err2 != nil {
			t.Fatal("Expected no error, got:", err2)
		}

		_, err1 := service.UpdateTask(Result{ID: item.ID, Status: Cancelled})
		if !errors.Is(err1, ErrInvalidTransaction) {
			t.Fatal("Expected ", ErrInvalidTransaction, "got:", err1)
		}
	})
}
