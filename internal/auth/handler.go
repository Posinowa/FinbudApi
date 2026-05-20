package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Posinowa/FinbudApp/internal/apperror"
	"github.com/Posinowa/FinbudApp/pkg/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", middleware.RegisterRateLimiter.Middleware(), h.Register)
		auth.POST("/login", middleware.LoginRateLimiter.Middleware(), h.Login)
		auth.POST("/google", middleware.LoginRateLimiter.Middleware(), h.GoogleLogin)
		auth.POST("/forgot-password", middleware.LoginRateLimiter.Middleware(), h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", middleware.AuthMiddleware(), h.Logout)
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	resp, statusCode, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	resp, statusCode, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("unauthorized", err.Error()))
		return
	}

	c.JSON(statusCode, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	resp, statusCode, err := h.service.Refresh(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("unauthorized", err.Error()))
		return
	}

	c.JSON(statusCode, resp)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	statusCode, err := h.service.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, gin.H{"message": "Sifre sifirlama e-postasi gonderildi"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	statusCode, err := h.service.ResetPassword(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, gin.H{"message": "Sifre basariyla sifirlandi"})
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	resp, statusCode, err := h.service.GoogleLogin(c.Request.Context(), req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("unauthorized", err.Error()))
		return
	}

	c.JSON(statusCode, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	// Authorization header'dan access token'ı al
	accessToken := ""
	if parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2); len(parts) == 2 {
		accessToken = parts[1]
	}

	statusCode, err := h.service.Logout(c.Request.Context(), req.RefreshToken, accessToken)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, gin.H{"message": "Basariyla cikis yapildi"})
}