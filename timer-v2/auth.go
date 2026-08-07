package timer

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/frame/fbr"
)

const (
	sessionCookieName = "timer_token"
	sessionMaxAge     = 7 * 24 * 3600 // 7 天
)

var (
	authPassword string
	sessionToken string
)

// deriveToken 从密码派生固定的会话 token (HMAC-SHA256)。
// 无状态、确定性：同一密码始终派生同一 token。
func deriveToken(password string) string {
	h := hmac.New(sha256.New, []byte(password))
	h.Write([]byte("timer-v2-session"))
	return hex.EncodeToString(h.Sum(nil))
}

// initAuth 读取密码并派生 token。密码为空时不启用认证（sessionToken 保持空）。
func initAuth() {
	authPassword = cfg.GetString("password", "")
	if authPassword == "" {
		authPassword = os.Getenv("PASSWORD")
	}
	if authPassword != "" {
		sessionToken = deriveToken(authPassword)
	}
}

// authMiddleware 认证中间件。
// 无密码 -> 放行；登录/登出接口 -> 放行；其余 -> 校验 Cookie。
// 注：fbr.Ctx 的 next() 为非导出方法，包外中间件用 fiber.Ctx 的 Next()，
// 并以 func(c fbr.Ctx) error 形式让 fbr transfer 的 dealErr 处理下游错误。
func authMiddleware(c fbr.Ctx) error {
	if sessionToken == "" || c.Path() == "/api/login" || c.Path() == "/api/logout" {
		return c.Next()
	}
	if c.Cookies(sessionCookieName) != sessionToken {
		c.Status(401)
		c.JSON(map[string]interface{}{"code": 401, "msg": "未登录"})
		return nil
	}
	return c.Next()
}

// loginHandler POST /api/login
func loginHandler(c fbr.Ctx) {
	req := &struct {
		Password string `json:"password"`
	}{}
	c.Parse(req)
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(authPassword)) != 1 {
		c.Status(401)
		c.JSON(map[string]interface{}{"code": 401, "msg": "密码错误"})
		return
	}
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   sessionMaxAge,
	})
	c.JSON(map[string]interface{}{"code": 200, "msg": "ok"})
}

// logoutHandler POST /api/logout
func logoutHandler(c fbr.Ctx) {
	c.ClearCookie(sessionCookieName)
	c.JSON(map[string]interface{}{"code": 200, "msg": "ok"})
}
