package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
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
	r = r.PathPrefix("/users").Subrouter()
	r.Use(h.authMiddleware.AuthMiddleware())
	r.HandleFunc("/", h.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/{id}", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/{id}", h.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/{id}", h.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/", h.ListUsers).Methods(http.MethodGet)
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

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("[Handler][CreateUser] invalid request body", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	user, err := h.userService.Create(r.Context(), input)
	if err != nil {
		h.log.Error("[Handler][CreateUser] failed to create user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusCreated,
		Message:    MessageCreateUserSuccess,
		Data:       user,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		h.log.Error("[Handler][GetUser] failed to get user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusNotFound,
			Message:    ErrNotFound,
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    MessageGetUserSuccess,
		Data:       user,
	})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var input models.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("[Handler][UpdateUser] invalid request body", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	user, err := h.userService.Update(r.Context(), id, input)
	if err != nil {
		h.log.Error("[Handler][UpdateUser] failed to update user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    MessageUpdateUserSuccess,
		Data:       user,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		h.log.Error("[Handler][DeleteUser] invalid request")
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}
	if err := h.userService.Delete(r.Context(), id); err != nil {
		h.log.Error("[Handler][DeleteUser] failed to delete user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    MessageDeleteUserSuccess,
		Data:       nil,
	})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// get offset and limit from query params
	page, err := util.ParseToInt(r.URL.Query().Get("page"))
	if err != nil {
		h.log.Error("[Handler][ListUsers] invalid query params", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}
	pageSize, err := util.ParseToInt(r.URL.Query().Get("pageSize"))
	if err != nil {
		h.log.Error("[Handler][ListUsers] invalid query params", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	if page < 1 {
		page = 1
	}

	if pageSize > 100 {
		pageSize = 100
	}

	users, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("[Handler][ListUsers] failed to list users", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrInternalServerError,
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    MessageListUserSuccess,
		Data:       users,
	})
}

func (h *UserHandler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	validateRequest := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&validateRequest); err != nil {
		h.log.Error("[Handler][ValidateUser] invalid request body", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    ErrInvalidRequest,
			Data:       nil,
		})
		return
	}

	user, err := h.userService.Validate(r.Context(), validateRequest.Email, validateRequest.Password)
	if err != nil {
		h.log.Error("[Handler][ValidateUser] failed to validate user", zap.Error(err))
		util.ResponseError(w, util.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Người dùng không hợp lệ",
			Data:       nil,
		})
		return
	}

	util.ResponseOK(w, util.Response{
		StatusCode: http.StatusOK,
		Message:    "Người dùng hợp lệ",
		Data: struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}{
			UserID: user.ID,
			Role:   user.Role,
		},
	})
}
