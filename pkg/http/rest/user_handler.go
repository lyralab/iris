package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/root-ali/iris/pkg/auth"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
)

var validate *validator.Validate

type UserSignupBody struct {
	UserName        string `json:"username" validate:"required,min=3,max=30"`
	FirstName       string `json:"firstname,omitempty" validate:"omitempty,min=3,max=30"`
	LastName        string `json:"lastname,omitempty" validate:"omitempty,min=3,max=30"`
	Password        string `json:"password" validate:"required,passwordStrength,min=8,max=30"`
	ConfirmPassword string `json:"confirm-password" validate:"required,eqfield=Password"`
	Mobile          string `json:"mobile_number,omitempty" validate:"omitempty,len=11,numeric"`
	Email           string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserUpdateBody struct {
	UserName  string `json:"username" validate:"required,min=3,max=30"`
	FirstName string `json:"firstname,omitempty" validate:"omitempty,min=3,max=30"`
	LastName  string `json:"lastname,omitempty" validate:"omitempty,min=3,max=30"`
	Password  string `json:"password,omitempty" validate:"omitempty,passwordStrength,min=8,max=30"`
	Mobile    string `json:"mobile,omitempty" validate:"omitempty,len=11,numeric"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserVerifyBody struct {
	UserName string `json:"username" validate:"required,min=3,max=30"`
}

type UserSigninBody struct {
	UserName string `json:"username" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,passwordStrength,min=8,max=30"`
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("passwordStrength", PasswordValidation)
}

func PasswordValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	specialChars := "@$!%*?&#^"

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	return len(password) >= 8 && hasUpper && hasLower && hasDigit && hasSpecial
}

func AddUserHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)

		logger.Infow("Received request body",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.Request.RemoteAddr,
			"body", string(bodyBytes),
		)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var u UserSignupBody
		decoder := json.NewDecoder(c.Request.Body)
		err := decoder.Decode(&u)
		if err != nil {
			var syntaxError *json.SyntaxError
			var unmarshalTypeError *json.UnmarshalTypeError

			switch {
			case errors.As(err, &syntaxError):
				logger.Errorw("Request body contains badly-formed JSON",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
					"offset", syntaxError.Offset,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Request body contains badly-formed JSON",
				})
			case errors.As(err, &unmarshalTypeError):
				logger.Errorw("Request body contains an invalid value for the "+unmarshalTypeError.Field+
					" field (at position "+fmt.Sprint(unmarshalTypeError.Offset)+")",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
					"field", unmarshalTypeError.Field,
					"type", unmarshalTypeError.Type,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"message": "Request body contains an invalid value for the " +
						unmarshalTypeError.Field + " field",
				})
			default:
				logger.Errorw("Invalid request body",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request body",
				})
			}
			return
		}

		err = validate.Struct(u)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make([]string, 0)
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s",
					e.Field(), e.Tag()))
			}
			logger.Errorw("Validation failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
				"body", string(bodyBytes),
				"errors", errorMessages,
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Validation failed",
				"errors":  errorMessages,
			})
			return
		}
		newUser := u.toUser()
		err = us.AddUser(newUser)
		logger.Info("error is ", err)

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Cannot add user right now: " + err.Error(),
				"status": "error",
			})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created", "username": u.UserName})
	}
}

func LoginUserHandler(us user.UserInterfaceService, aths auth.AuthServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var u UserSigninBody
		decoder := json.NewDecoder(c.Request.Body)
		err := decoder.Decode(&u)

		if err != nil {
			var syntaxError *json.SyntaxError
			var unmarshalTypeError *json.UnmarshalTypeError

			switch {
			case errors.As(err, &syntaxError):
				logger.Errorw("Request body contains badly-formed JSON",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
					"offset", syntaxError.Offset,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Request body contains badly-formed JSON",
				})
			case errors.As(err, &unmarshalTypeError):
				logger.Errorw("Request body contains an invalid value for the "+unmarshalTypeError.Field+
					" field (at position "+fmt.Sprint(unmarshalTypeError.Offset)+")",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
					"field", unmarshalTypeError.Field,
					"type", unmarshalTypeError.Type,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"message": "Request body contains an invalid value for the " +
						unmarshalTypeError.Field + " field",
				})
			default:
				logger.Errorw("Invalid request body",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"remote_addr", c.Request.RemoteAddr,
					"body", string(bodyBytes),
					"error", err,
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request body",
				})
			}
			return
		}

		err = validate.Struct(u)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make([]string, 0)
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s",
					e.Field(), e.Tag()))
			}
			logger.Errorw("Validation failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
				"body", string(bodyBytes),
				"errors", errorMessages,
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Validation failed",
				"errors":  errorMessages,
			})
			return
		}
		reqUser := user.User{}
		reqUser.UserName = u.UserName
		reqUser.Password = u.Password
		err = us.ValidateUser(&reqUser)
		if errors.Is(err, iris_error.ErrPasswordNotMatch) {
			logger.Errorw("Wrong password is entered ", "user", u.UserName)
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "error",
				"message": "Cannot continue the process",
			})
			return
		} else if err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"status":  "error",
				"message": "Internal Server Error",
			})
			return
		}
		err = us.GetUserRole(&reqUser)
		if err != nil {
			logger.Errorw("cannot get user role", "user", reqUser.UserName)
		}
		logger.Infow("user role is ", "role", reqUser.Role, "username", reqUser.UserName)
		token, err := aths.GenerateToken(reqUser.UserName, reqUser.Role)
		if err != nil {
			logger.Errorw("cannot generate token for user",
				"user", reqUser.UserName, "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		//c.Cookie("jwt")
		c.JSON(200, gin.H{"status": "OK", "token": token})
	}
}

func VerifyUserHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var u UserVerifyBody
		decoder := json.NewDecoder(c.Request.Body)
		err := decoder.Decode(&u)
		if err != nil {
			logger.Errorw("cannot decode body", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"status": "error", "error": err})
			return
		}
		userVerify := user.User{}
		userVerify.UserName = u.UserName
		err = us.VerifyUser(&userVerify)
		if err != nil {
			logger.Errorw("cannot verify user", "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		c.JSON(200, gin.H{"status": "OK"})
	}
}

func UpdateUserHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var u UserUpdateBody
		decoder := json.NewDecoder(c.Request.Body)
		err := decoder.Decode(&u)
		if err != nil {
			logger.Errorw("cannot decode body", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"status": "error", "error": err})
			return
		}
		err = validate.Struct(u)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make([]string, 0)
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s",
					e.Field(), e.Tag()))
			}
			logger.Errorw("Validation failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
				"body", string(bodyBytes),
				"errors", errorMessages,
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Validation failed",
				"errors":  errorMessages,
			})
			return
		}
		updateUser := u.toUser()
		logger.Infow("Updating user data",
			"user_id", updateUser.ID,
			"username", updateUser.UserName,
			"mail", updateUser.Email,
			"mobile", updateUser.Mobile)
		err = us.UpdateUser(updateUser)
		if err != nil {
			logger.Errorw("cannot update user", "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		c.JSON(200, gin.H{"status": "OK", "username": updateUser.UserName})
	}
}

func GetAllUsersHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := us.GetAllUsers()
		if err != nil {
			logger.Errorw("cannot get all users", "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		count := len(users)
		userResponse := make([]map[string]interface{}, count)
		for i, u := range users {
			userResponse[i] = map[string]interface{}{
				"id":        u.ID,
				"username":  u.UserName,
				"firstName": u.FirstName,
				"lastName":  u.LastName,
				"email":     u.Email,
				"mobile":    u.Mobile,
				"status":    u.Status,
			}
		}
		c.JSON(200, gin.H{"status": "OK", "users": userResponse, "count": count})
	}
}

func GetUserInfoHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName, ok := c.Get("username")
		if !ok {
			logger.Errorw("cannot get username or role from context")
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": "token is invalid"})
			return
		}

		u, err := us.GetByUserName(userName.(string))
		if err != nil {
			logger.Errorw("cannot get user", "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		userResponse := map[string]string{
			"username":  u.UserName,
			"firstName": u.FirstName,
			"lastName":  u.LastName,
			"email":     u.Email,
			"mobile":    u.Mobile,
		}
		c.JSON(200, gin.H{"status": "OK", "user": userResponse, "user_id": u.ID})
	}
}

func GetUserByIDHandler(us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		u, err := us.GetByUserId(userID)
		if err != nil {
			logger.Errorw("cannot get user", "error", err)
			c.AbortWithStatusJSON(500, gin.H{"status": "error", "error": err})
			return
		}
		userResponse := map[string]string{
			"username":  u.UserName,
			"firstName": u.FirstName,
			"lastName":  u.LastName,
			"email":     u.Email,
			"mobile":    u.Mobile,
		}
		c.JSON(200, gin.H{"status": "OK", "user": userResponse, "user_id": u.ID})
	}
}

func (ub *UserSignupBody) toUser() *user.User {
	return &user.User{
		UserName:  ub.UserName,
		FirstName: ub.FirstName,
		LastName:  ub.LastName,
		Password:  ub.Password,
		Email:     ub.Email,
	}
}

func (ub *UserUpdateBody) toUser() *user.User {
	u := &user.User{
		UserName: ub.UserName, // UserName is always required
	}

	// Only set fields that are not empty
	if ub.FirstName != "" {
		u.FirstName = ub.FirstName
	}
	if ub.LastName != "" {
		u.LastName = ub.LastName
	}
	if ub.Password != "" {
		u.Password = ub.Password
	}
	if ub.Mobile != "" {
		u.Mobile = ub.Mobile
	}
	if ub.Email != "" {
		u.Email = ub.Email
	}

	return u
}
