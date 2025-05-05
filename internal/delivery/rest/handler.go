package rest

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
	"user_service/api/middleware"
	"user_service/internal/models"
	"user_service/internal/service"
	"user_service/internal/util"
)

type UserHandler struct {
	userService    service.UserService
	authMiddleware *middleware.AuthMiddleware
	log            *zap.Logger
}

func NewUserHandler(userService service.UserService, log *zap.Logger, authMiddleware *middleware.AuthMiddleware) *UserHandler {
	return &UserHandler{userService: userService, log: log, authMiddleware: authMiddleware}
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to the User Service!"))
	})
	r = r.PathPrefix("/users").Subrouter()
	r.Use(h.authMiddleware.AuthMiddleware())
	r.Use(h.authMiddleware.OwnerMiddleware())
	r.Handle("", h.authMiddleware.ACLMiddleware("admin", "user")((http.HandlerFunc)(h.ListUsers))).Methods(http.MethodGet)
	r.Handle("/", h.authMiddleware.ACLMiddleware("admin")((http.HandlerFunc)(h.CreateUser))).Methods(http.MethodPost)
	r.Handle("/{id}", h.authMiddleware.ACLMiddleware("admin", "user")((http.HandlerFunc)(h.GetUser))).Methods(http.MethodGet)
	r.Handle("/{id}", h.authMiddleware.ACLMiddleware("admin", "user")((http.HandlerFunc)(h.UpdateUser))).Methods(http.MethodPut)
	r.Handle("/{id}", h.authMiddleware.ACLMiddleware("admin")((http.HandlerFunc)(h.DeleteUser))).Methods(http.MethodDelete)
	r.Handle("/{id}/change-password", h.authMiddleware.ACLMiddleware("admin", "user")((http.HandlerFunc)(h.ChangePassword))).Methods(http.MethodPut)
}

var (
	ErrInvalidRequest      = "Dữ liệu không hợp lệ, vui lòng kiểm tra lại"
	ErrInternalServerError = "Lỗi hệ thống, vui lòng thử lại sau"
	ErrNotFound            = "Không tìm thấy, vui lòng kiểm tra lại"
)

var (
	MessageCreateUserSuccess = "Tạo user thành công"
	MessageGetUserSuccess    = "Lấy thông tin user thành công"
	MessageUpdateUserSuccess = "Cập nhật thông tin user thành công"
	MessageDeleteUserSuccess = "Xóa user thành công"
	MessageListUserSuccess   = "Lấy danh sách user thành công"
)

// CreateUser godoc
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Security JWT
// @Param create_user body models.CreateUserInput true "Create user request"
// @Success      201  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("[Handler][CreateUser] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	user, err := h.userService.Create(r.Context(), input)
	if err != nil {
		h.log.Error("[Handler][CreateUser] failed to create user", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseOK(w, user, http.StatusCreated)
}

// GetUser godoc
// @Summary Get user
// @Description Get user information
// @Tags users
// @Accept json
// @Produce json
// @Security JWT
// @Param id path string true "User ID"
// @Success      200  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      404  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.log.Info("[Handler][GetUser] user not found", zap.Error(err))
			util.ResponseOK(w, nil, http.StatusOK)
			return
		}
		h.log.Info("[Handler][GetUser] failed to get user", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseOK(w, user, http.StatusOK)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Security JWT
// @Param id path string true "User ID"
// @Param update_user body models.UpdateUserInput true "Update user request"
// @Success      200  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var input models.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("[Handler][UpdateUser] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	user, err := h.userService.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.log.Error("[Handler][UpdateUser] user not found", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    BAD_REQUEST,
				TimeStamp: time.Now().String(),
				Message:   ErrInvalidRequest,
			}, http.StatusBadRequest)
			return
		}

		h.log.Error("[Handler][UpdateUser] failed to update user", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseOK(w, user, http.StatusOK)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user
// @Tags users
// @Accept json
// @Produce json
// @Security JWT
// @Param id path string true "User ID"
// @Success      200  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		h.log.Error("[Handler][DeleteUser] invalid request")
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.log.Info("[Handler][DeleteUser] user not found", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    BAD_REQUEST,
				TimeStamp: time.Now().String(),
				Message:   ErrInvalidRequest,
			}, http.StatusBadRequest)
			return
		}

		h.log.Error("[Handler][DeleteUser] failed to delete user", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseOK(w, util.ResponseSuccess{
		Message: "Xóa user thành công",
	}, http.StatusOK)
}

// ListUsers godoc
// @Summary List users
// @Description List users
// @Tags users
// @Accept json
// @Produce json
// @Security JWT
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success      200  {object}  util.Response
// @Failure      400  {object}  util.Response
// @Failure      500  {object}  util.Response
// @Router       /users/list [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// get offset, limit, search username, password, filter role, status from query params
	params := util.GetPaginationParams(r)

	if role := r.URL.Query().Get("role"); role != "" {
		params.Filters["role"] = role
	}

	if status := r.URL.Query().Get("status"); status != "" {
		params.Filters["status"] = status
	}

	users, count, err := h.userService.List(r.Context(), params)
	if err != nil {
		if !errors.Is(err, service.ErrUserNotFound) {
			h.log.Error("[Handler][ListUsers] failed to list users", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    INTERNAL_SERVER_ERROR,
				TimeStamp: time.Now().String(),
				Message:   ErrInvalidRequest,
			}, http.StatusInternalServerError)
		}
		return
	}
	if users == nil {
		users = make([]*models.User, 0)
	}
	util.ResponseOK(w, util.CreatePaginationResponse(users, count, params), http.StatusOK)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		h.log.Error("[Handler][DeleteUser] invalid request")
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	var input models.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("[Handler][ChangePassword] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	if err := input.Validate(); err != nil {
		h.log.Error("[Handler][ChangePassword] invalid request body", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    BAD_REQUEST,
			TimeStamp: time.Now().String(),
			Message:   ErrInvalidRequest,
		}, http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePassword(r.Context(), id, input); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.log.Info("[Handler][ChangePassword] user not found", zap.Error(err))
			util.ResponseErr(w, util.ResponseError{
				Status:    BAD_REQUEST,
				TimeStamp: time.Now().String(),
				Message:   ErrInvalidRequest,
			}, http.StatusBadRequest)
			return
		}

		h.log.Error("[Handler][ChangePassword] failed to change password", zap.Error(err))
		util.ResponseErr(w, util.ResponseError{
			Status:    INTERNAL_SERVER_ERROR,
			TimeStamp: time.Now().String(),
			Message:   ErrInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseOK(w, util.ResponseSuccess{
		Message: "Đổi mật khẩu thành công",
	}, http.StatusOK)

}
