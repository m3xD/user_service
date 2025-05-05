package rest

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
	"user_service/internal/models"
	"user_service/internal/service"
	"user_service/internal/util"
)

type AuthHandler struct {
	authService models.AuthService
	userService service.UserService
	jwtService  util.JwtImpl
	router      *mux.Router
	log         *zap.Logger
}

func NewAuthHandler(authService models.AuthService, userService service.UserService, router *mux.Router, log *zap.Logger, jwtService util.JwtImpl) *AuthHandler {
	return &AuthHandler{authService: authService, log: log, router: router, userService: userService, jwtService: jwtService}
}

func (h *AuthHandler) RegisterRoutes() {
	h.router = h.router.PathPrefix("/auth").Subrouter()
	h.router.HandleFunc("/login", h.Login).Methods("POST")
	h.router.HandleFunc("/register", h.Register).Methods("POST")
	h.router.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	h.router.HandleFunc("/logout", h.Logout).Methods("POST")
	//h.router.HandleFunc("/forgot-password", h.ForgotPassword).Methods("POST")
	//h.router.HandleFunc("/reset-password", h.ResetPassword).Methods("POST")
}

var (
	MessageLoginSuccess         = "Đăng nhập thành công"
	MessageRegisterSuccess      = "Đăng ký thành công"
	MessageWrongEmailOrPassword = "Email hoặc mật khẩu không đúng"
	MessageUserExists           = "User đã tồn tại"
	MessageRefreshTokenExpired  = "Refresh token hết hạn"
	MessageRefreshTokenSuccess  = "Refresh token thành công"
	MessageInvalidCredentials   = "Thông tin đăng nhập không hợp lệ"
)

const (
	BAD_REQUEST           = "BAD_REQUEST"
	UNAUTHORIZED          = "UNAUTHORIZED"
	INTERNAL_SERVER_ERROR = "INTERNAL_SERVER_ERROR"
)

// Login godoc
// @Summary Login
// @Description Login user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login request"
// @Success      200  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      404  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var loginRequest models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		h.log.Error("[Handler][Login] failed to parse request", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusBadRequest)
		return
	}

	err = loginRequest.Validate()
	if err != nil {
		var field, reason string
		if errors.Is(err, models.ErrEmailEmpty) {
			field, reason = "email", "email is required"
		} else {
			field, reason = "password", "password is required"
		}
		h.log.Error("[Handler][Login] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors: []util.ErrReason{
				{
					Field:   field,
					Message: reason,
				},
			},
		}, http.StatusBadRequest)
		return
	}

	// Call service
	res, err := h.userService.Validate(r.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.log.Error("[Handler][Login] user not found", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    UNAUTHORIZED,
				TimeStamp: time.Now().String(),
				Message:   MessageWrongEmailOrPassword,
				Errors:    nil,
			}, http.StatusUnauthorized)
			return
		}
		h.log.Error("[Handler][Login] failed to login", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateAccessToken(res.ID, res.Role)
	if err != nil {
		h.log.Error("[Handler][Login] failed to generate token", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwtService.GenerateRefreshToken(res.ID, res.Role)
	if err != nil {
		h.log.Error("[AuthService][Login] failed to generate refresh token", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Store refresh token
	err = h.authService.SaveToken(r.Context(), refreshToken, res.ID)
	if err != nil {
		h.log.Error("[AuthService][Login] failed to save refresh token", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Response
	util.ResponseOK(w, models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: &models.UserSummary{
			Id:    res.ID,
			Name:  res.FullName,
			Email: res.Email,
			Role:  res.Role,
		},
	}, http.StatusOK)
}

// Register godoc
// @Summary Register
// @Description Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param register body models.RegisterRequest true "Register request"
// @Success      201  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var registerRequest models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerRequest)
	if err != nil {
		h.log.Error("[Handler][Register] failed to parse request", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
			Errors:    nil,
		}, http.StatusBadRequest)
		return
	}

	err = registerRequest.Validate()
	if err != nil {
		h.log.Error("[Handler][Register] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
			Errors:    nil,
		}, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	user, err := h.userService.GetByEmail(r.Context(), registerRequest.Email)

	if user != nil {
		h.log.Error("[Handler][Register] user already exists", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageUserExists,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	} else if err != nil && !errors.Is(err, service.ErrUserNotFound) {
		h.log.Error("[Handler][Register] failed to check user", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Call service
	user, err = h.userService.Create(r.Context(), models.CreateUserInput{
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
		FullName: registerRequest.FullName,
		Phone:    "default",
		Role:     models.RoleUser,
	})

	if err != nil {
		h.log.Error("[Handler][Register] failed to register", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Response
	util.ResponseOK(w, struct {
		Message string              `json:"message"`
		User    *models.UserSummary `json:"user"`
	}{
		Message: MessageRegisterSuccess,
		User: &models.UserSummary{
			Id:    user.ID,
			Name:  user.FullName,
			Email: user.Email,
			Role:  user.Role,
		},
	}, http.StatusCreated)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var refreshTokenRequest models.RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&refreshTokenRequest)
	if err != nil {
		h.log.Error("[Handler][RefreshToken] failed to parse request", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	}

	if refreshTokenRequest.RefreshToken == "" {
		h.log.Error("[Handler][RefreshToken] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	}

	// Call service
	accessToken, refreshToken, err := h.authService.RefreshToken(r.Context(), refreshTokenRequest.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrExpiredToken) {
			h.log.Error("[Handler][RefreshToken] token expired", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    UNAUTHORIZED,
				TimeStamp: time.Now().String(),
				Message:   MessageInvalidCredentials,
				Errors:    nil,
			}, http.StatusUnauthorized)
			return
		}
		h.log.Error("[Handler][RefreshToken] failed to refresh token", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
			Errors:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// Response
	util.ResponseOK(w, map[string]interface{}{
		"token":        accessToken,
		"refreshToken": refreshToken,
	}, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var refreshReq models.RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&refreshReq)
	if err != nil {
		h.log.Error("[Handler][Logout] failed to parse request", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	}

	if refreshReq.RefreshToken == "" {
		h.log.Error("[Handler][Logout] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	}

	err = h.authService.Logout(r.Context(), refreshReq.RefreshToken)
	if err != nil {
		h.log.Error("[Handler][Logout] failed to logout", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    UNAUTHORIZED,
			TimeStamp: time.Now().String(),
			Message:   MessageInvalidCredentials,
			Errors:    nil,
		}, http.StatusUnauthorized)
		return
	}

	util.ResponseOK(w, util.ResponseSuccess{
		Message: "Operation successful",
	}, http.StatusOK)
}
