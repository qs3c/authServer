package main

import (
	"authServer/server/internal/repository"
	"authServer/server/internal/repository/dao"
	"authServer/server/internal/service"
	"authServer/server/internal/web"
	"authServer/server/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()

	server := initWebServer()
	initUserHdl(db, server)
	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/auth"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type"},
		//AllowHeaders: []string{"content-type"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	login := &middleware.LoginMiddlewareBuilder{}

	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())

	return server
}
