package auth

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected_error error
	}{
		{name: "upper and lowercase letters", input: "Chipotle", expected_error: nil},
		{name: "numbers", input: "chipotle2", expected_error: nil},
		{name: "valid special characters", input: "chipotle_groo.ve", expected_error: nil},
		{name: "space in Username", input: "chipotle groove", expected_error: userInvalidCharErr},
		{name: "invalid special characters", input: "chipotle@", expected_error: userInvalidCharErr},
		{name: "Username below minimum length", input: "a", expected_error: userMinimumLengthErr},
		{name: "Username equal to minimum length", input: "aaa", expected_error: nil},
		{name: "Username above maximum length", input: "imsuperlonglikeyouhavenoideahowlongiamtruly", expected_error: userMaximumLengthErr},
		{name: "Username equal to maximum length", input: "11111111111111111", expected_error: nil},
		{name: "valid Username length", input: "chipotle", expected_error: nil},
		{name: "empty Username", input: "", expected_error: userMinimumLengthErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validateUsername(test.input)

			assertError(t, result, test.expected_error)
		})
	}
}
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected_error error
	}{
		{name: "valid password", input: "Abc12345!", expected_error: nil},
		{name: "no uppercase", input: "abc12345!", expected_error: passInvalidCharErr},
		{name: "no special char", input: "Abc12345", expected_error: passInvalidCharErr},
		{name: "no number", input: "Abcasdasasd!", expected_error: passInvalidCharErr},
		{name: "no lowercase", input: "A123123123123!", expected_error: passInvalidCharErr},
		{name: "no letter", input: "123123123123!", expected_error: passInvalidCharErr},
		{name: "below minimum", input: "ola", expected_error: passMinLengthErr},
		{name: "minimum", input: "Abc1234!", expected_error: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validatePassword(test.input)

			assertError(t, got, test.expected_error)
		})
	}
}

func TestPasswordCrypt(t *testing.T) {
	password := "Abc123!"
	result, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	if bcrypt.CompareHashAndPassword(result, []byte(password)) != nil {
		t.Error("password hashes do not match")
	}
}

func TestRegister(t *testing.T) {
	users := UserDatabase{
		UsersByEmail:    map[string]*User{},
		UsersByUsername: map[string]*User{},
	}
	tests := []struct {
		name           string
		input          [3]string
		expected_error error
	}{
		{name: "valid register", input: [3]string{"mail@gmail.com", "testestestestes", "Abc12345!"}, expected_error: nil},
		{name: "invalid Username", input: [3]string{"mail2@gmail.com", "test test", "Abc12345!"}, expected_error: userInvalidCharErr},
		{name: "invalid password", input: [3]string{"mail3@gmail.com", "testestestes", "Abc123!"}, expected_error: passMinLengthErr},
		{name: "invalid email", input: [3]string{"hiimnotvalid", "testestestestes", "Abc12345!"}, expected_error: invalidEmailErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registered_user, err := users.RegisterUser(test.input[0], test.input[1], test.input[2])

			if err != test.expected_error {
				t.Fatalf("unexpected error, got %q, expected %q", err, test.expected_error)
			}

			user_in_db, _ := users.getUser(test.input[0])

			if user_in_db != nil {
				assertUsers(t, registered_user, *user_in_db)
				assertUUID(t, user_in_db.Id)
			} else if test.expected_error == nil {
				t.Fatalf("expected a user, but got nil")
			}
		})
	}

	t.Run("existing user", func(t *testing.T) {
		user, err := users.RegisterUser("mail4@gmail.com", "testestestestes", "Abc12345!")

		if err != usernameRegisteredErr {
			t.Fatalf("unexpected error, got %q, expected %q", err, usernameRegisteredErr)
		}

		assertUsers(t, user, User{})
	})

	t.Run("existing email", func(t *testing.T) {
		user, err := users.RegisterUser("mail@gmail.com", "hiiiiiiiiii", "Abc12345!")

		if err != emailRegisteredErr {
			t.Fatalf("unexpected error, got %q, expected %q", err, emailRegisteredErr)
		}

		assertUsers(t, user, User{})
	})
}

func TestGetUser(t *testing.T) {
	test_user := User{Email: "mail@gmail.com", Username: "testestsetsetst", Password: "test"}
	users := UserDatabase{
		UsersByEmail:    map[string]*User{test_user.Email: &test_user},
		UsersByUsername: map[string]*User{test_user.Username: &test_user},
	}
	tests := []struct {
		name           string
		input          string
		expected_user  User
		expected_error error
	}{
		{name: "existing email", input: "mail@gmail.com", expected_user: test_user, expected_error: nil},
		{name: "non-existing user", input: "idontexist", expected_user: User{}, expected_error: userNotFoundErr},
		{name: "existing Username", input: "testestsetsetst", expected_user: test_user, expected_error: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user_in_db, err := users.getUser(test.input)

			if err != test.expected_error {
				t.Fatalf("unexpected error, got %q, expected %q", err, test.expected_error)
			}

			if user_in_db != nil {
				assertUsers(t, *user_in_db, test.expected_user)
			} else if test.expected_error == nil {
				t.Fatalf("expected a user, but got nil")
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected_error error
	}{
		{name: "valid email", input: "mail@gmail.com", expected_error: nil},
		{name: "valid email 2", input: "mail@gmail", expected_error: nil},
		{name: "valid email 3", input: "mail@gmail.com (Tester Testerino)", expected_error: nil},
		{name: "invalid email", input: "thisisdefinitelynotanemail", expected_error: invalidEmailErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validEmailErr := validateEmail(test.input)

			assertError(t, validEmailErr, test.expected_error)
		})
	}
}

func TestGenerateUUID(t *testing.T) {
	user := User{Email: "test", Username: "test", Password: "A"}

	generateUUID(&user)

	assertUUID(t, user.Id)

}

func TestLogin(t *testing.T) {
	users := UserDatabase{
		UsersByEmail:    map[string]*User{},
		UsersByUsername: map[string]*User{},
	}
	test_user, _ := users.RegisterUser("mail@gmail.com", "testertester", "Abc12345!")
	tests := []struct {
		name           string
		input          [2]string
		expected_error error
		expected_user  User
	}{
		{name: "valid login with email", input: [2]string{"mail@gmail.com", "Abc12345!"}, expected_error: nil, expected_user: test_user},
		{name: "valid login with Username", input: [2]string{"testertester", "Abc12345!"}, expected_error: nil, expected_user: test_user},
		{name: "nonexisting email", input: [2]string{"mail2@gmail.com", "test12345!"}, expected_error: userNotFoundErr, expected_user: User{}},
		{name: "nonexisting Username", input: [2]string{"test2", "test12345!"}, expected_error: userNotFoundErr, expected_user: User{}},
		{name: "wrong password", input: [2]string{"mail@gmail.com", "wrong1235!"}, expected_error: wrongPasswordErr, expected_user: User{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logged_user, log_err := LogIn(users, test.input[0], test.input[1])

			if test.expected_error != log_err {
				t.Fatalf("unexpected error, expected %q, got %q", test.expected_error, log_err)
			}

			assertUsers(t, logged_user, test.expected_user)
		})
	}

}

//helpers

func assertError(t testing.TB, actual_error, expected_error error) {
	t.Helper()
	if actual_error != expected_error {
		t.Errorf("got %q, expected %q", actual_error, expected_error)
	}
}

func assertUsers(t testing.TB, actual_user, expected_user User) {
	t.Helper()
	if !reflect.DeepEqual(actual_user, expected_user) {
		t.Errorf("got %v, expected %v", actual_user, expected_user)
	}
}

func assertUUID(t testing.TB, actual_UUID uuid.UUID) {
	t.Helper()
	if actual_UUID == uuid.Nil {
		t.Errorf("id unchanged: %v", actual_UUID)
	}
}
