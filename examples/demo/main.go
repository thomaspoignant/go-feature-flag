package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

var users = make(map[string]ffcontext.EvaluationContext, 2500)

func main() {
	configFile := flag.String("configFile", "./demo-flags.goff.yaml", "flags.goff.yaml")
	flag.Parse()

	_ = ffclient.Init(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: *configFile,
		},
	})

	e := echo.New()
	e.Static("/js", "js")
	e.Static("/css", "css")
	// Instantiate a template registry and register all html files inside the view folder
	e.Renderer = &TemplateRegistry{templates: template.Must(template.ParseGlob("view/*.html"))}

	// init users
	for i := 0; i < 2500; i++ {
		id := uuid.New()
		u := ffcontext.NewEvaluationContext(id.String())
		users[fmt.Sprintf("user%d", i)] = u
	}

	e.GET("/", apiHandler)
	if err := (echo.StartConfig{Address: ":8080", HideBanner: true}).Start(context.Background(), e); err != nil {
		log.Fatal(err)
	}
}

type TemplateRegistry struct {
	templates *template.Template
}

func (t *TemplateRegistry) Render(c *echo.Context, w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func apiHandler(c *echo.Context) error {
	mapToRender := make(map[string]string, 2500)
	for k, user := range users {
		color, _ := ffclient.StringVariation("color-box", user, "grey")
		mapToRender[k] = color
	}
	return c.Render(http.StatusOK, "template.html", mapToRender)
}
