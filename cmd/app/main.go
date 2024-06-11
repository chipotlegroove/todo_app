package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"todo_app/pkg/auth"
	"todo_app/pkg/tasks"

	"github.com/google/uuid"
)

func main() {
	var userInput string
	var email string
	var username string
	var password string
	var id string
	reader := bufio.NewReader(os.Stdin)

	//USERS PREP
	users := auth.UserDatabase{
		UsersByEmail:    make(map[string]*auth.User),
		UsersByUsername: make(map[string]*auth.User),
	}

	//read users file
	users_file, fileErr := os.OpenFile("../data/users.csv", os.O_RDWR|os.O_CREATE, 0666)

	if fileErr != nil {
		log.Println("users file could not be opened")
	}

	defer users_file.Close()

	//load users into db
	userCsvReader := csv.NewReader(users_file)
	userCsvWriter := csv.NewWriter(users_file)

	defer userCsvWriter.Flush()

	for {
		rec, err := userCsvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		user := auth.User{Id: uuid.MustParse(rec[0]), Email: rec[1], Username: rec[2], Password: rec[3]}
		users.UsersByEmail[user.Email] = &user
		users.UsersByUsername[user.Username] = &user
	}

	//TASKS PREP
	UserTasks := make(map[uuid.UUID]tasks.TaskList)

	//read tasks file
	tasks_file, fileErr := os.Open("../data/tasks.csv")

	if fileErr != nil {
		log.Println("tasks file could not be opened")
	}

	defer tasks_file.Close()

	//load tasks into database
	taskCsvReader := csv.NewReader(tasks_file)

	for {
		rec, err := taskCsvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		userId := uuid.MustParse(rec[0])
		id, convErr := strconv.Atoi(rec[1])

		if convErr != nil {
			log.Fatal(err)
		}
		task := tasks.Task{Id: id, Name: rec[2], Description: rec[3], Date: rec[4], TaskStatus: rec[5]}

		taskList, found := UserTasks[userId]
		if found {
			taskList[id] = &task
		} else {
			newTaskList := make(tasks.TaskList)
			newTaskList[id] = &task
			UserTasks[userId] = newTaskList
		}
	}

	//write task file

	defer func() {
		tasks_file, fileErr = os.OpenFile("../data/tasks.csv", os.O_WRONLY|os.O_TRUNC, 0644)
		if fileErr != nil {
			log.Fatal("tasks file could not be open")
		}
		taskCsvWriter := csv.NewWriter(tasks_file)
		for userId, taskList := range UserTasks {
			for taskId, task := range taskList {
				stringId := strconv.Itoa(taskId)
				taskData := []string{userId.String(), stringId, strings.TrimSpace(task.Name), strings.TrimSpace(task.Description), strings.TrimSpace(task.Date), strings.TrimSpace(task.TaskStatus)}
				err := taskCsvWriter.Write(taskData)
				if err != nil {
					log.Fatal("couldnt write task to tasks file")
				}
			}
		}
		taskCsvWriter.Flush()
	}()

outer:
	for {
		fmt.Println("welcome to the best to do list app")
		fmt.Println("1.- Register")
		fmt.Println("2.- Login")
		fmt.Println("3.- Exit")
		_, err := fmt.Scanln(&userInput)
		if err != nil {
			log.Fatal(err)
		}
		switch userInput {
		case "1":
			fmt.Println("Please enter your email")
			_, emailErr := fmt.Scanln(&email)
			if emailErr != nil {
				fmt.Println(emailErr)
				continue
			}
			fmt.Println("Please enter your username")
			_, usernameErr := fmt.Scanln(&username)
			if usernameErr != nil {
				fmt.Println(usernameErr)
				continue
			}
			fmt.Println("Please enter your password (Password must have at least 8 characters, and must contain at least one uppercase and lowercase letter, one number and one of the following symbols: @$!%*?&)")
			_, passErr := fmt.Scanln(&password)
			if passErr != nil {
				fmt.Println(passErr)
				continue
			}
			user, registerErr := users.RegisterUser(email, username, password)
			if registerErr != nil {
				fmt.Println(registerErr)
				continue
			}
			userData := []string{user.Id.String(), user.Email, user.Username, user.Password}
			err := userCsvWriter.Write(userData)
			if err != nil {
				fmt.Println("error writing to file")
				continue
			}
		options_menu:
			for {
				fmt.Printf("Welcome %q, what would you like to do today\n", user.Username)
				fmt.Println("1.- Add task")
				fmt.Println("2.- See all tasks")
				fmt.Println("3.- Edit task")
				fmt.Println("4.- Delete task")
				fmt.Println("5.- Mark task as complete")
				fmt.Println("6.- Log out")
				_, err := fmt.Scanln(&userInput)
				if err != nil {
					fmt.Println(err)
				}
				switch userInput {
				case "1":
					for {
						fmt.Println("Enter the name of the task:")
						taskName, nameErr := reader.ReadString('\n')
						if nameErr != nil {
							fmt.Println(nameErr)
						}
						fmt.Println("Enter the description of the task:")
						taskDesc, descErr := reader.ReadString('\n')
						if descErr != nil {
							fmt.Println(descErr)
						}
						fmt.Println("Enter the date when you want to complete the task:")
						taskDate, dateErr := reader.ReadString('\n')
						if dateErr != nil {
							fmt.Println(dateErr)
						}
						taskDate = strings.TrimSpace(taskDate)
						loggedUserTasks, found := UserTasks[user.Id]
						if !found {
							loggedUserTasks = make(tasks.TaskList)
							UserTasks[user.Id] = loggedUserTasks
						}
						newTask, err := loggedUserTasks.AddTask(taskName, taskDesc, taskDate)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Printf("Succesfully added new task:%v", newTask)
							continue options_menu
						}
					}
				case "2":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
					fmt.Println("Your tasks")
					table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
					fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
					for _, task := range loggedUserTasks {
						fmt.Fprintln(table, task.String())
					}
					table.Flush()
				case "3":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				edit_menu:
					for {
						fmt.Println("Enter the number of the task you want to edit, the field you want to change and the new value. (example: '2 name New Name')")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						editInput, editErr := reader.ReadString('\n')
						if editErr != nil {
							fmt.Println(editErr)
							continue edit_menu
						}
						fields := strings.Fields(editInput)
						if len(fields) < 3 {
							fmt.Println("Please enter an appropiate input")
							continue edit_menu
						}
						editId, err := strconv.Atoi(fields[0])
						if err != nil {
							fmt.Println("Please enter a valid input for the task number")
							continue edit_menu
						}
						editFieldName := fields[1]
						editNewValue := strings.Join(fields[2:], " ")
						_, err = loggedUserTasks.UpdateField(editId, editFieldName, editNewValue)
						if err != nil {
							fmt.Println(err)
							continue edit_menu
						}
						continue options_menu
					}
				case "4":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				delete_menu:
					for {
						fmt.Println("Select the number of the task you wish to delete or 0 to return to the previous menu")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						taskId, idErr := reader.ReadString('\n')
						if idErr != nil {
							fmt.Println(idErr)
							continue delete_menu
						}
						numId, err := strconv.Atoi(strings.TrimSpace(taskId))
						if err != nil {
							fmt.Println(err)
							continue delete_menu
						}
						if numId == 0 {
							continue options_menu
						}
						fmt.Printf("Deleting task #%q, type Y to confirm, any other input to cancel\n", strings.TrimSpace(taskId))
						confirm, confErr := reader.ReadString('\n')
						if confErr != nil {
							fmt.Println(confErr)
							continue delete_menu
						}
						confirm = strings.ToLower(strings.TrimSpace(confirm))
						if confirm == "y" {
							err := loggedUserTasks.DeleteTask(numId)
							if err != nil {
								fmt.Println(err)
								continue delete_menu
							}
							continue options_menu
						} else {
							continue delete_menu
						}

					}
				case "5":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				complete_menu:
					for {
						fmt.Println("Select the number of the task you wish to complete or 0 to return to the previous menu")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						taskId, idErr := reader.ReadString('\n')
						if idErr != nil {
							fmt.Println(idErr)
							continue complete_menu
						}
						numId, err := strconv.Atoi(strings.TrimSpace(taskId))
						if err != nil {
							fmt.Println(err)
							continue complete_menu
						}
						if numId == 0 {
							continue options_menu
						}
						fmt.Printf("Completing task #%q, type Y to confirm, any other input to cancel\n", strings.TrimSpace(taskId))
						confirm, confErr := reader.ReadString('\n')
						if confErr != nil {
							fmt.Println(confErr)
							continue complete_menu
						}
						confirm = strings.ToLower(strings.TrimSpace(confirm))
						if confirm == "y" {
							err := loggedUserTasks.CompleteTask(numId)
							if err != nil {
								fmt.Println(err)
								continue complete_menu
							}
							continue options_menu
						} else {
							continue complete_menu
						}
					}
				case "6":
					continue outer
				default:
					fmt.Println("u stupid")
				}
			}

		case "2":
			fmt.Println("Please enter your username or email")
			_, idErr := fmt.Scanln(&id)
			if idErr != nil {
				log.Println(idErr)
				continue
			}
			fmt.Println("Please enter your password")
			_, passErr := fmt.Scanln(&password)
			if passErr != nil {
				log.Println(passErr)
				continue
			}
			user, logErr := auth.LogIn(users, id, password)
			if logErr != nil {
				log.Println(logErr)
				continue
			}
		logged_options_menu:
			for {
				fmt.Printf("Welcome %q, what would you like to do today\n", user.Username)
				fmt.Println("1.- Add task")
				fmt.Println("2.- See all tasks")
				fmt.Println("3.- Edit task")
				fmt.Println("4.- Delete task")
				fmt.Println("5.- Mark task as complete")
				fmt.Println("6.- Log out")
				_, err := fmt.Scanln(&userInput)
				if err != nil {
					log.Fatal(err)
				}
				switch userInput {
				case "1":
					for {
						fmt.Println("Enter the name of the task:")
						taskName, nameErr := reader.ReadString('\n')
						if nameErr != nil {
							fmt.Println(nameErr)
						}
						fmt.Println("Enter the description of the task:")
						taskDesc, descErr := reader.ReadString('\n')
						if descErr != nil {
							fmt.Println(descErr)
						}
						fmt.Println("Enter the date when you want to complete the task:")
						taskDate, dateErr := reader.ReadString('\n')
						if dateErr != nil {
							fmt.Println(dateErr)
						}
						taskDate = strings.TrimSpace(taskDate)
						loggedUserTasks, found := UserTasks[user.Id]
						if !found {
							loggedUserTasks = make(tasks.TaskList)
							UserTasks[user.Id] = loggedUserTasks
						}
						newTask, err := loggedUserTasks.AddTask(taskName, taskDesc, taskDate)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Printf("Succesfully added new task:%v", newTask)
							continue logged_options_menu
						}
					}
				case "2":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
					fmt.Println("Your tasks")
					table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
					fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
					for _, task := range loggedUserTasks {
						fmt.Fprintln(table, task.String())
					}
					table.Flush()
				case "3":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				logged_edit_menu:
					for {
						fmt.Println("Enter the number of the task you want to edit, the field you want to change and the new value. (example: '2 name New Name')")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						editInput, editErr := reader.ReadString('\n')
						if editErr != nil {
							fmt.Println(editErr)
							continue logged_edit_menu
						}
						fields := strings.Fields(editInput)
						if len(fields) < 3 {
							fmt.Println("Please enter an appropiate input")
							continue logged_edit_menu
						}
						editId, err := strconv.Atoi(fields[0])
						if err != nil {
							fmt.Println("Please enter a valid input for the task number")
							continue logged_edit_menu
						}
						editFieldName := fields[1]
						editNewValue := strings.Join(fields[2:], " ")
						_, err = loggedUserTasks.UpdateField(editId, editFieldName, editNewValue)
						if err != nil {
							fmt.Println(err)
							continue logged_edit_menu
						}
						continue logged_options_menu
					}
				case "4":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				logged_delete_menu:
					for {
						fmt.Println("Select the number of the task you wish to delete or 0 to return to the previous menu")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						taskId, idErr := reader.ReadString('\n')
						if idErr != nil {
							fmt.Println(idErr)
							continue logged_delete_menu
						}
						numId, err := strconv.Atoi(strings.TrimSpace(taskId))
						if err != nil {
							fmt.Println(err)
							continue logged_delete_menu
						}
						if numId == 0 {
							continue logged_options_menu
						}
						fmt.Printf("Deleting task #%q, type Y to confirm, any other input to cancel\n", strings.TrimSpace(taskId))
						confirm, confErr := reader.ReadString('\n')
						if confErr != nil {
							fmt.Println(confErr)
							continue logged_delete_menu
						}
						confirm = strings.ToLower(strings.TrimSpace(confirm))
						if confirm == "y" {
							err := loggedUserTasks.DeleteTask(numId)
							if err != nil {
								fmt.Println(err)
								continue logged_delete_menu
							}
							continue logged_options_menu
						} else {
							continue logged_delete_menu
						}

					}

				case "5":
					loggedUserTasks, found := UserTasks[user.Id]
					if !found {
						loggedUserTasks = make(tasks.TaskList)
						UserTasks[user.Id] = loggedUserTasks
					}
				logged_complete_menu:
					for {
						fmt.Println("Select the number of the task you wish to complete or 0 to return to the previous menu")
						table := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
						fmt.Fprintln(table, "Task Number\t", "Name\t", "Description\t", "Date\t", "Task Status\t")
						for _, task := range loggedUserTasks {
							fmt.Fprintln(table, task.String())
						}
						table.Flush()
						taskId, idErr := reader.ReadString('\n')
						if idErr != nil {
							fmt.Println(idErr)
							continue logged_complete_menu
						}
						numId, err := strconv.Atoi(strings.TrimSpace(taskId))
						if err != nil {
							fmt.Println(err)
							continue logged_complete_menu
						}
						if numId == 0 {
							continue logged_options_menu
						}
						fmt.Printf("Completing task #%q, type Y to confirm, any other input to cancel\n", strings.TrimSpace(taskId))
						confirm, confErr := reader.ReadString('\n')
						if confErr != nil {
							fmt.Println(confErr)
							continue logged_complete_menu
						}
						confirm = strings.ToLower(strings.TrimSpace(confirm))
						if confirm == "y" {
							err := loggedUserTasks.CompleteTask(numId)
							if err != nil {
								fmt.Println(err)
								continue logged_complete_menu
							}
							continue logged_options_menu
						} else {
							continue logged_complete_menu
						}

					}
				case "6":
					continue outer
				default:
					fmt.Println("u stupid")
				}
			}
		case "3":
			fmt.Println("cya")
			return
		default:
			fmt.Println("u stupid")
		}
	}

}
