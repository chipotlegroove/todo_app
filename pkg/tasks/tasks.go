package tasks

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TaskNotFoundErr = TaskError("Task not found")
	TaskDateErr     = TaskError("Date format is invalid")
	TaskNameErr     = TaskError("Task cannot have an empty name")
	InvalidFieldErr = TaskError("The selected field does not exist")
)

type UserTaskList map[uuid.UUID]TaskList

type Task struct {
	Id          int
	Name        string
	Description string
	Date        string
	TaskStatus  string
}

type TaskError string

func (err TaskError) Error() string {
	return string(err)
}

type TaskList map[int]*Task

func (tasks TaskList) GetTask(id int) (*Task, error) {
	task, found := tasks[id]
	if found {
		return task, nil
	}
	return nil, TaskNotFoundErr
}

func (tasks TaskList) AddTask(name, description, date string) (Task, error) {
	if name != "" {
		_, err := time.Parse("02-01-2006", date)
		if err == nil {
			newId := len(tasks) + 1
			task := Task{Id: newId, Name: name, Description: description, Date: date, TaskStatus: "pending"}
			tasks[newId] = &task
			return task, nil
		} else {
			return Task{}, TaskDateErr
		}

	} else {
		return Task{}, TaskNameErr
	}
}

func (tasks TaskList) DeleteTask(id int) error {
	_, err := tasks.GetTask(id)
	taskFound := err == nil
	if taskFound {
		delete(tasks, id)
		return nil
	}
	return err
}

func (tasks TaskList) UpdateField(id int, field, new_value string) (Task, error) {
	formattedField := strings.ToLower(strings.TrimSpace(field))
	task, err := tasks.GetTask(id)
	taskFound := err == nil
	if taskFound {
		switch formattedField {
		case "name", "1":
			task.Name = new_value
			return *task, nil
		case "description", "2":
			task.Description = new_value
			return *task, nil
		case "date", "3":
			_, err := time.Parse("02-01-2006", new_value)
			if err == nil {
				task.Date = new_value
				return *task, nil
			}
			return Task{}, TaskDateErr

		default:
			return Task{}, InvalidFieldErr
		}
	}
	return Task{}, TaskNotFoundErr
}

func (tasks TaskList) CompleteTask(id int) error {
	task, err := tasks.GetTask(id)
	taskFound := err == nil
	if taskFound {
		task.TaskStatus = "complete"
	}
	return err
}

func (t Task) String() string {
	return fmt.Sprintf("%v.-\t%q \t %q \t %q \t %q", strconv.Itoa(t.Id), strings.TrimSpace(t.Name), t.Description, t.Date, t.TaskStatus)
}
