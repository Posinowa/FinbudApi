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

	// Web redirect sayfası — e-postadaki HTTPS linkten uygulamayı açar
	r.GET("/reset-password", h.ResetPasswordPage)
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

func (h *Handler) ResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Geçersiz bağlantı")
		return
	}

	deepLink := "finbud://reset-password?token=" + token
	html := `<!DOCTYPE html>
<html lang="tr">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Finbud - Şifre Sıfırlama</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
           background: #f5f6fa; display: flex; align-items: center;
           justify-content: center; min-height: 100vh; padding: 24px; }
    .card { background: #fff; border-radius: 16px; padding: 40px 32px;
            max-width: 400px; width: 100%; text-align: center;
            box-shadow: 0 4px 24px rgba(0,0,0,0.08); }
    .icon { font-size: 48px; margin-bottom: 16px; }
    h1 { font-size: 22px; font-weight: 700; color: #2D3748; margin-bottom: 8px; }
    p { font-size: 14px; color: #718096; margin-bottom: 32px; line-height: 1.6; }
    .btn { display: block; background: #4F5D75; color: #fff; padding: 16px 32px;
           border-radius: 12px; text-decoration: none; font-size: 16px;
           font-weight: 600; margin-bottom: 16px; }
    .note { font-size: 12px; color: #A0AEC0; }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">🔐</div>
    <h1>Şifre Sıfırlama</h1>
    <p>Finbud uygulamasında yeni şifrenizi oluşturmak için aşağıdaki butona dokunun.</p>
    <a class="btn" href="` + deepLink + `">Uygulamada Aç</a>
    <p class="note">Bu bağlantı 1 saat geçerlidir.</p>
  </div>
  <script>
    // Otomatik yönlendirmeyi dene
    window.location.href = "` + deepLink + `";
  </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
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