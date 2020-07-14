package main

import (
	"ferry"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Login struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func main() {
	validate := validator.New()
	config := ferry.Config{
		Validator: validate,
	}
	app := ferry.InitServer(&config)

	notFoundHandler := func(ctx *ferry.Ctx) error {
		return ctx.Redirect(http.StatusMovedPermanently, "https://google.com")
	}

	app.HandleNotFound(notFoundHandler)

	//app.Use(middleware.Gzip())

	//app.Use(func(ctx *ferry.Ctx) error {
	//	fmt.Println("hey!, this is middleware")
	//	return ctx.Next()
	//})
	// un comment below code to see early response from middleware
	//app.Use(func(ctx *ferry.Ctx) error {
	//	fmt.Println("Early response from middleware")
	//	return ctx.Send(http.StatusOK, "From app middleware")
	//})

	//app.Get("/", func(ctx *ferry.Ctx) error {
	//	return ctx.Send(http.StatusOK, "Hello, World!")
	//})

	app.Get("/download", func(ctx *ferry.Ctx) error {
		return ctx.SendAttachment("server.go", "server.go")
	})

	app.Post("/upload", func(ctx *ferry.Ctx) error {
		return ctx.UploadFile("static/index.html", "index")
	})

	app.Get("/login", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "Hello, World!, This is login")
	})

	app.Get("/hey", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "Heey!")
	})

	app.Put("/heyput", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "heyPut")
	})

	app.Delete("/heydelete", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "heyDelete")
	})

	app.Post("/login", func(ctx *ferry.Ctx) error {
		login := Login{}
		err := ctx.Bind(&login)
		if err != nil {
			return err
		}
		return ctx.Json(http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Welcome %s", login.Username),
		})
	})

	app.Get("/hello", func(ctx *ferry.Ctx) error {
		params := ctx.GetQueryParams()
		return ctx.Json(http.StatusOK, params)
	})

	app.Get("/hello/single", func(ctx *ferry.Ctx) error {
		params := ctx.GetQueryParam("key")
		return ctx.Send(http.StatusOK, fmt.Sprintf("key is %s", params))
	})
	// group routing
	auth := app.Group("/auth")
	auth.Delete("/authdelete", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "authdelete")
	})
	auth.Put("/authput", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "authput")
	})
	auth.Use(func(ctx *ferry.Ctx) error {
		//fmt.Println("this is auth level middleware")
		return ctx.Next()
	})

	//// un comment below code for early response from middleware
	//auth.Use(func(ctx *ferry.Ctx) error {
	//	fmt.Println("this is auth level middleware2")
	//	return ctx.Send(http.StatusOK, "response from auth middleware")
	//})

	auth.Get("/:name", func(ctx *ferry.Ctx) error {
		name := ctx.GetParam("name")
		return ctx.Send(http.StatusOK, fmt.Sprintf("Welcome %s", name))
	})

	auth.Get("/:name/lol", func(ctx *ferry.Ctx) error {
		name := ctx.GetParam("name")
		return ctx.Send(http.StatusOK, fmt.Sprintf("Registered %s with lol", name))
	})

	auth.Get("/:name/:age", func(ctx *ferry.Ctx) error {
		params := ctx.GetParams()
		return ctx.Json(http.StatusOK, params)
	})

	dashBoard := app.Group("/dashboard")
	dashBoard.Use(func(ctx *ferry.Ctx) error {
		//fmt.Println("dashboard")
		return ctx.Next()
	})
	dashBoard.Get("/all", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "all")
	})
	// nested routes
	dashboardPrivate := dashBoard.Group("/private")
	dashboardPrivate.Use(func(ctx *ferry.Ctx) error {
		//fmt.Println("dashboard private")
		return ctx.Next()
	})
	dashboardPrivate.Get("/all", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "private all")
	})

	app.ServeFile("/js", "static/js/login.js")
	app.ServerDir("/", "static")

	log.Fatal(app.Listen("localhost:3000"))
}
