package server

import (
	"strconv"

	contro "bbyd/internal/controllers"
	mdware "bbyd/internal/controllers/middleware"
	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/logs"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Run(s config.Server) error {
	e := echo.New()
	routes(e)
	adr := ":" + strconv.Itoa(s.Port)
	err := e.Start(adr)
	if err != nil {
		logs.Error("server start failed at "+adr+". ", zap.Error(err))
		return err
	}
	return nil
}

func routes(e *echo.Echo) {
	e.Use(echomw.Logger())

	path := config.Configs.Constants.ApiPathRoot
	api := e.Group(path, mdware.UseResponseContext)
	{
		user := api.Group("/user")
		{
			user.GET("/p/:name", contro.UserIndexHandler, mdware.TokenVerify) // user index view
			user.POST("/p", contro.RegisterHandler)                           // user register
			user.PUT("/p/:name", contro.SetinfoHandler, mdware.TokenVerify)   // change user info
			user.DELETE("/p/:name", contro.DeleteHandler, mdware.TokenVerify) // delete user

			user.GET("/token", contro.LoginHandler)                         // login
			user.POST("/token/email/:name", contro.LoginEmailSendHandler)   // login email send
			user.GET("/token/email/", contro.CodeLoginHandler)              // login by email code
			user.DELETE("/token", contro.LogoutHandler, mdware.TokenVerify) // logout

			// user.GET("/post", contro.UserPostViewHandler, mdware.TokenVerify)  // view one's posts
		}

		node := api.Group("/node")
		{
			node.GET("", contro.NodeIndexHandler, mdware.CanHaveToken)            // view all the nodes
			node.POST("", contro.CreateNodeHandler, mdware.TokenVerify)           // create a new node, admin only
			node.PUT("/:nodeid", contro.UpdateNodeHandler, mdware.TokenVerify)    // update node info, admin only
			node.DELETE("/:nodeid", contro.DeleteNodeHandler, mdware.TokenVerify) // delete a node, admin only
			node.GET("/:nodeid", contro.ForumIndexHandler, mdware.CanHaveToken)   // view all the posts in a node
		}

		post := api.Group("/post")
		{
			post.GET("/:postid", contro.PostViewHandler, mdware.CanHaveToken)     // view certain post
			post.POST("", contro.CreatePostHandler, mdware.TokenVerify)           // create a new post
			post.PUT("/:postid", contro.UpdatePostHandler, mdware.TokenVerify)    // update a post
			post.DELETE("/:postid", contro.DeletePostHandler, mdware.TokenVerify) // delete a post
		}
	}
}
