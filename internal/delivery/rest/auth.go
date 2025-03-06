package rest

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"user_service/internal/models"
	"user_service/internal/service"
	"user_service/internal/util"
)

type AuthHandler struct {
	authService models.AuthService
	userService service.UserService
	router      *mux.Router
	log         *zap.Logger
}

func NewAuthHandler(authService models.AuthService, userService service.UserService, router *mux.Router, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, log: log, router: router, userService: userService}
}

func (h *AuthHandler) RegisterRoutes() {
	h.router = h.router.PathPrefix("/auth").Subrouter()
	h.router.HandleFunc("/login", h.Login).Methods("POST")
	h.router.HandleFunc("/register", h.Register).Methods("POST")
}

var (
	MessageLoginSuccess    = "Đăng nhập thành công"
	MessageRegisterSuccess = "Đăng ký thành công"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var loginRequest models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		h.log.Error("[Handler][Login] failed to parse request", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	err = loginRequest.Validate()
	if err != nil {
		h.log.Error("[Handler][Login] invalid request body", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	// Call service
	res, err := h.userService.Validate(r.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		h.log.Error("[Handler][Login] failed to login", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	// Response
	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    MessageLoginSuccess,
		Data:       res,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var registerRequest models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerRequest)
	if err != nil {
		h.log.Error("[Handler][Register] failed to parse request", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	err = registerRequest.Validate()
	if err != nil {
		h.log.Error("[Handler][Register] invalid request body", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	// Check if user already exists
	user, err := h.userService.GetByEmail(r.Context(), registerRequest.Email)

	if user != nil {
		h.log.Error("[Handler][Register] user already exists", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "User already exists",
			Data:       nil,
		})
		return
	} else if err != nil && !errors.Is(err, service.ErrUserNotFound) {
		h.log.Error("[Handler][Register] failed to check user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
	}

	// Call service
	user, err = h.userService.Create(r.Context(), models.CreateUserInput{
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
		FullName: registerRequest.FullName,
		Phone:    registerRequest.Phone,
		Role:     models.RoleUser,
	})

	if err != nil {
		h.log.Error("[Handler][Register] failed to register", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
		return
	}

	// Response
	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusCreated,
		Message:    MessageRegisterSuccess,
		Data:       user,
	})
}
