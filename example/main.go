package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ferry-go/ferry"

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

	app.Get("/", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "Hello, World")
	})

	// redirect to new url
	app.Get("/redirect", func(ctx *ferry.Ctx) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/static")
	})

	app.Get("/name/:name", func(ctx *ferry.Ctx) error {
		name := ctx.GetParam("name")
		return ctx.Send(http.StatusOK, fmt.Sprintf("hello, %s", name))
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

	// Grouping your route
	auth := app.Group("/auth")
	auth.Get("/login", func(ctx *ferry.Ctx) error {
		return ctx.Send(http.StatusOK, "Hello, World")
	})

	// path and dir name
	app.ServerDir("/static", "static")

	// single file
	app.ServeFile("/js", "static/login.js")

	// Downloading file
	app.Get("/download", func(ctx *ferry.Ctx) error {
		return ctx.SendAttachment("static/login.js", "login.js")
	})

	// uploading file
	app.Post("/upload", func(ctx *ferry.Ctx) error {
		return ctx.UploadFile("static/login.js", "login.js")
	})

	log.Fatal(app.Listen("localhost:3000"))
}
