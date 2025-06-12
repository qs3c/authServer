package web

import (
	"authServer/server/internal/domain"
	"authServer/server/internal/service"
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
	// GET /users/check_times  路径参数 Params
	//ug.GET("/check_times/:auth_type", h.CheckAuthTimes)
	ug.GET("/check_times", h.CheckAuthTimes)
	// GET /users/minus_one 请求参数 QueryParams
	ug.GET("/minus_one", h.MinusOneAuthTimes)

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
	u, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		// 登录成功了设置服务器的 session
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}

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
		AuthKey          string `json:"auth_key"`
		Email            string `json:"email"`
		AuthHideTimes    string `json:"auth_hide_times"`
		AuthExtractTimes string `json:"auth_extract_times"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		fmt.Println("bind failed!")
		return
	}
	AuthHideTimes, err := strconv.Atoi(req.AuthHideTimes)
	if err != nil {
		ctx.String(http.StatusOK, "输入授权次数违规")
		return
	}

	AuthExtractTimes, err := strconv.Atoi(req.AuthExtractTimes)
	if err != nil {
		ctx.String(http.StatusOK, "输入授权次数违规")
		return
	}

	fmt.Println(req.Email, AuthHideTimes, AuthExtractTimes)
	// 2.
	if req.AuthKey != authKey {
		ctx.String(http.StatusOK, "授权密钥错误!")
		return
	}

	err = h.svc.InitAuthTimes(ctx, req.Email, AuthHideTimes, AuthExtractTimes)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.String(http.StatusOK, "账号不存在")
		} else {
			ctx.String(http.StatusOK, "失败！")
		}
		return
	}
	ctx.String(http.StatusOK, "成功！")
}

func (h *UserHandler) CheckAuthTimes(ctx *gin.Context) {
	//authType := ctx.Param("auth_type")

	var userId int64
	value, exist := ctx.Get("userId")
	if !exist {
		ctx.String(http.StatusOK, "获取用户信息失败！")
		return
	}
	userId = value.(int64)

	hideRemainTimes, extractRemainTimes, err := h.svc.CheckAuthTimes(ctx, userId)

	if err != nil {
		ctx.String(http.StatusOK, "<UNK>")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"hideRemainTimes":    hideRemainTimes,
		"extractRemainTimes": extractRemainTimes,
	})
}

func (h *UserHandler) MinusOneAuthTimes(ctx *gin.Context) {

	authType := ctx.Query("auth_type")
	fmt.Println("auth_type:", authType)

	var AuthTypeBool bool
	if authType == "h" {
		AuthTypeBool = false
	} else if authType == "e" {
		AuthTypeBool = true
	} else {
		ctx.String(http.StatusOK, "h/e 类型异常！")
		return
	}

	var userId int64
	value, exist := ctx.Get("userId")
	if !exist {
		ctx.String(http.StatusOK, "获取用户信息失败！")
		return
	}
	userId = value.(int64)

	remainTimes, err := h.svc.MinusOneAuthTimes(ctx, userId, AuthTypeBool)

	if err != nil {
		ctx.String(http.StatusOK, "<UNK>")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"remainTimes": remainTimes,
	})
	//ctx.String(http.StatusOK, "%d", strconv.Itoa(remainTimes))
}
