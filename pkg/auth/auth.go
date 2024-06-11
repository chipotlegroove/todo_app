package auth

import (
	"net/mail"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	userRegex        = regexp.MustCompile("^[a-zA-Z0-9_.]+$")
	lowercaseRegex   = regexp.MustCompile(`[a-z]`)
	uppercaseRegex   = regexp.MustCompile(`[A-Z]`)
	digitRegex       = regexp.MustCompile(`\d`)
	specialCharRegex = regexp.MustCompile(`[@$!%*?&]`)
	passwordRules    = []*regexp.Regexp{
		lowercaseRegex,
		uppercaseRegex,
		digitRegex,
		specialCharRegex,
		specialCharRegex,
	}
)

const (
	usernameMaxLength     = 17
	usernameMinLength     = 3
	passMinLength         = 8
	userMinimumLengthErr  = UsernameErr("Username must be at least 3 characters long")
	userMaximumLengthErr  = UsernameErr("Username must be below 17 characters long")
	userInvalidCharErr    = UsernameErr("Username cannot contain spaces or special characters beside commas and dots")
	passMinLengthErr      = PasswordErr("Password must have at least 8 characters")
	passInvalidCharErr    = PasswordErr("Password must contain at least one uppercase and lowercase letter, one number and one of the following symbols: @$!%*?&")
	userNotFoundErr       = UserErr("User not found")
	usernameRegisteredErr = UserErr("This username has already been registered")
	invalidEmailErr       = EmailErr("Your email address is not in a valid format")
	emailRegisteredErr    = UserErr("This email has already been registered")
	wrongPasswordErr      = PasswordErr("Incorrect password")
)

type User struct {
	Id       uuid.UUID
	Email    string
	Username string
	Password string
}

type UserDatabase struct {
	UsersByEmail    map[string]*User
	UsersByUsername map[string]*User
}

func (users UserDatabase) getUser(id string) (*User, error) {
	user_by_email, user_found := users.UsersByEmail[id]
	if user_found {
		return user_by_email, nil
	}
	user_by_username, user_found := users.UsersByUsername[id]
	if user_found {
		return user_by_username, nil
	}
	return nil, userNotFoundErr
}

func (users UserDatabase) RegisterUser(email, username, password string) (User, error) {
	validEmailError := validateEmail(email)
	if validEmailError != nil {
		return User{}, validEmailError
	}

	validUsernameError := validateUsername(username)
	if validUsernameError != nil {
		return User{}, validUsernameError
	}
	validPasswordError := validatePassword(password)
	if validPasswordError != nil {
		return User{}, validPasswordError
	}
	_, errByUsername := users.getUser(username)
	_, errByEmail := users.getUser(email)

	usernameAvailable := errByUsername != nil
	emailAvailable := errByEmail != nil

	if emailAvailable {
		if usernameAvailable {
			hashed_pass, pass_error := bcrypt.GenerateFromPassword([]byte(password), 12)
			if pass_error == nil {
				newUser := User{Email: email, Username: username, Password: string(hashed_pass)}
				generateUUID(&newUser)
				users.UsersByUsername[username] = &newUser
				users.UsersByEmail[email] = &newUser
				return newUser, nil
			}
			return User{}, pass_error
		}
		return User{}, usernameRegisteredErr
	}
	return User{}, emailRegisteredErr

}

type PasswordErr string

func (e PasswordErr) Error() string {
	return string(e)
}

type UsernameErr string

func (e UsernameErr) Error() string {
	return string(e)
}

type UserErr string

func (e UserErr) Error() string {
	return string(e)
}

type EmailErr string

func (e EmailErr) Error() string {
	return string(e)
}

func validateUsername(username string) error {
	if len(username) < usernameMinLength {
		return userMinimumLengthErr
	}
	if len(username) > usernameMaxLength {
		return userMaximumLengthErr
	}
	if !userRegex.MatchString(username) {
		return userInvalidCharErr
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < passMinLength {
		return passMinLengthErr
	}
	for _, rule := range passwordRules {
		if !rule.MatchString(password) {
			return passInvalidCharErr
		}
	}
	return nil
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)

	if err != nil {
		return invalidEmailErr
	}

	return nil
}

func generateUUID(user *User) {
	user.Id = uuid.New()
}

func LogIn(users UserDatabase, id, password string) (loggedUser User, err error) {
	user, err := users.getUser(id)
	if err == nil {
		correct_password := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
		if correct_password {
			return *user, nil
		}
		return User{}, wrongPasswordErr
	}
	return User{}, userNotFoundErr
}
