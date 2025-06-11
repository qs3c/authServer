package web

import (
	"authServer/server/internal/domain"
	"authServer/server/internal/service"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	authKey              = "fC0lL0tF5bX7wR6jI7nN6mA5zM7bU7iZ"
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	//server.POST("/user", h.SignUp)
	//server.PUT("/user", h.SignUp)
	//server.GET("/users/:username", h.Profile)
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	ug.POST("/login", h.Login)
	// POST /users/init_auth_times
	ug.POST("/init_auth_times", h.InitAuthTimes)
	// GET /users/use_hide
	ug.GET("/use_hide", h.UseHide)
	// GET /users/use_hide
	ug.GET("/use_extract", h.UseExtract)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		fmt.Println("debug 输出")
		fmt.Println(req.Email)
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	_, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		//sess := sessions.Default(ctx)
		//sess.Set("userId", u.Id)
		//sess.Options(sessions.Options{
		//	// 十五分钟
		//	MaxAge: 900,
		//})
		//err = sess.Save()
		//if err != nil {
		//	ctx.String(http.StatusOK, "系统错误")
		//	return
		//}
		fmt.Println("here?")
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) InitAuthTimes(ctx *gin.Context) {
	// 1.使用权限校验
	type Req struct {
		AuthKey            string `json:"auth_key"`
		Email              string `json:"email"`
		Auth_hide_times    int    `json:"auth_hide_times"`
		Auth_extract_times int    `json:"auth_extract_times"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 2.
	if req.AuthKey != authKey {
		return
	}

	err := h.svc.Init_auth_times(ctx, req.Email, req.Auth_hide_times, req.Auth_extract_times)
	if err != nil {
		ctx.String(http.StatusOK, "失败！")
		return
	}
	ctx.String(http.StatusOK, "成功！")
}

func (h *UserHandler) UseHide(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是 profile")
}

func (h *UserHandler) UseExtract(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是 profile")
}
