package tasks

import (
	"reflect"
	"testing"
)

func TestViewTask(t *testing.T) {
	sampleTask := Task{Id: 1, Name: "Make a test function", Description: "Jesse we need to test", Date: "31/05/2024", TaskStatus: "pending"}
	taskList := TaskList{sampleTask.Id: &sampleTask}

	t.Run("existing task", func(t *testing.T) {
		test_task, err := taskList.GetTask(1)

		if err == TaskNotFoundErr {
			t.Fatalf("got unexpected error, expected %q, got %q", err, TaskNotFoundErr)
		}

		AssertTask(t, *test_task, sampleTask)
	})

	t.Run("nonexisting task", func(t *testing.T) {
		no_task, err := taskList.GetTask(2)

		if err != TaskNotFoundErr {
			t.Fatalf("got unexpected error, expected %q, got %q", TaskNotFoundErr, err)
		}
		if no_task != nil {
			t.Errorf("expected task to be nil, is %q", no_task)
		}
	})

}

func TestAddTask(t *testing.T) {
	taskList := TaskList{}
	tests := []struct {
		name           string
		input          [3]string
		expected_error error
	}{
		{name: "valid task", input: [3]string{"Make a test function", "Jesse we need to test", "20-03-2014"}, expected_error: nil},
		{name: "empty name", input: [3]string{"", "lol", "2-06-2024"}, expected_error: TaskNameErr},
		{name: "invalid date", input: [3]string{"yea", "lol", "wtf wrong date?"}, expected_error: TaskDateErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addedTask, err := taskList.AddTask(test.input[0], test.input[1], test.input[2])

			if err != test.expected_error {
				t.Fatalf("got unexpected error, expected %q, got %q", test.expected_error, err)
			}

			if err == nil {
				if addedTask.Name != test.input[0] || addedTask.Description != test.input[1] || addedTask.Date != test.input[2] {
					t.Error("task fields not stored accurately")
				}
			}

		})
	}
}

func TestDeleteTask(t *testing.T) {
	sampleTask := Task{Id: 1, Name: "Make a test function", Description: "Jesse we need to test", Date: "31/05/2024", TaskStatus: "pending"}
	taskList := TaskList{sampleTask.Id: &sampleTask}

	t.Run("delete existing task", func(t *testing.T) {
		deleteErr := taskList.DeleteTask(1)

		if deleteErr != nil {
			t.Fatalf("expecting no errors got %q", deleteErr)
		}

		_, err := taskList.GetTask(1)

		if err == nil {
			t.Error("expected an error, got nothing")
		}
	})

	t.Run("delete nonexisting task", func(t *testing.T) {
		err := taskList.DeleteTask(420)

		if err != TaskNotFoundErr {
			t.Errorf("unexpected error, got %q, expected %q", err, TaskNotFoundErr)
		}

	})
}

func TestUpdateField(t *testing.T) {
	sampleTask := Task{Id: 1, Name: "Make a test function", Description: "Jesse we need to test", Date: "31/05/2024", TaskStatus: "pending"}
	taskList := TaskList{sampleTask.Id: &sampleTask}
	tests := []struct {
		name           string
		input          [2]string
		expected_error error
	}{
		{name: "no fields given", input: [2]string{"", ""}, expected_error: InvalidFieldErr},
		{name: "updating name", input: [2]string{"name", "new name"}, expected_error: nil},
		{name: "updating description", input: [2]string{"description", "jesse we tested"}, expected_error: nil},
		{name: "updating date", input: [2]string{"date", "08-10-2024"}, expected_error: nil},
		{name: "updating name mixed case", input: [2]string{"namE", "new name"}, expected_error: nil},
		{name: "updating description mixed case", input: [2]string{"descriptIon", "jesse we tested"}, expected_error: nil},
		{name: "updating date mixed case", input: [2]string{"dAte", "08-10-2024"}, expected_error: nil},
		{name: "untrimmed field name", input: [2]string{"dAte ", "08-10-2024"}, expected_error: nil},
		{name: "update date with number", input: [2]string{"3", "08-10-2024"}, expected_error: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := taskList.UpdateField(1, test.input[0], test.input[1])

			if err != test.expected_error {
				t.Fatalf("unexpected error, got %q, expected %q", err, test.expected_error)
			}
			if err == nil {
				task_in_db, err := taskList.GetTask(1)
				if err != nil {
					t.Fatal("expected to get task, got nothing")
				}
				switch test.input[0] {
				case "name", "1":
					if task_in_db.Name != test.input[1] {
						t.Fatalf("field did not take new value, expected name to be %q, it is %q", test.input[1], task_in_db.Name)
					}
				case "description", "2":
					if task_in_db.Description != test.input[1] {
						t.Fatalf("field did not take new value, expected name to be %q, it is %q", test.input[1], task_in_db.Description)
					}
				case "date", "3":
					if task_in_db.Date != test.input[1] {
						t.Fatalf("field did not take new value, expected name to be %q, it is %q", test.input[1], task_in_db.Date)
					}
				}
			}
		})
	}

	t.Run("update nonexisting task", func(t *testing.T) {
		_, err := taskList.UpdateField(3, "lol", "a")
		if err != TaskNotFoundErr {
			t.Fatalf("unexpected error, expected %q, got %q", TaskNotFoundErr, err)
		}
	})
}

//helpers

func AssertTask(t testing.TB, actualTask, expectedTask Task) {
	t.Helper()
	if !reflect.DeepEqual(actualTask, expectedTask) {
		t.Errorf("fetched wrong task, expected %v, got %v", expectedTask, actualTask)
	}

}
