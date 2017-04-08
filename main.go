package main

import (
	"html/template"

	"gopkg.in/kataras/iris.v6"

	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
)

// a custom Iris event policy, which will run when server interruped (i.e control+C)
// receives a func() error, most of packages are compatible with that on their Close/Shutdown/Cancel funcs.
func releaser(r func() error) iris.EventPolicy {
	return iris.EventPolicy{
		Interrupted: func(app *iris.Framework) {
			if err := r(); err != nil {
				app.Log(iris.ProdMode, "error while releasing resources: "+err.Error())
			}
		}}
}

// main ...
func main() {
	app := iris.New()

	db := NewDB("shortener.db")

	factory := NewFactory(DefaultGenerator, db)

	app.Adapt(
		// print all kind of errors and logs at os.Stdout
		iris.DevLogger(),
		// use the httprouter, you can use adpaotrs/gorillamux if you want
		httprouter.New(),
		// serve the "./templates" directory's "*.html" files with the HTML std view engine.
		view.HTML("./templates", ".html").Reload(true),
		// `db.Close` is a `func() error` so it can be a `releaser` too.
		// Wrap the db.Close with the releaser in order to be released when app exits or control+C
		// You probably never saw that before, clever pattern which I am able to use only with Iris :)
		releaser(db.Close),
	)

	app.Adapt(iris.TemplateFuncsPolicy{"isPositive": func(n int) bool {
		if n > 0 {
			return true
		}
		return false
	}})

	app.StaticWeb("/static", "./resources")

	app.Get("/", func(ctx *iris.Context) {
		ctx.MustRender("index.html", iris.Map{"url_count": db.Len()})
	})

	// find and execute a short url by its key
	// used on http://localhost:8080/u/dsaoj41u321dsa
	execShortURL := func(ctx *iris.Context, key string) {
		if key == "" {
			ctx.EmitError(iris.StatusBadRequest)
			return
		}

		value := db.Get(key)
		if value == "" {
			ctx.SetStatusCode(iris.StatusNotFound)
			ctx.Writef("Short URL for key: '%s' not found", key)
			return
		}

		ctx.Redirect(value, iris.StatusTemporaryRedirect)
	}

	app.Get("/u/:shortkey", func(ctx *iris.Context) {
		execShortURL(ctx, ctx.Param("shortkey"))
	})

	app.Get("/all", func(ctx *iris.Context) {
		var allShort = db.GetAll()
		var htmlStr string
		for k, v := range allShort {
			htmlStr += "<div><a target='_new' href='" + v + "'>" + k + " - " + v + "</a></div>"
		}

		ctx.MustRender("all.html", iris.Map{"shortens": template.HTML(htmlStr)})
	})

	app.Post("/shorten", func(ctx *iris.Context) {
		data := make(map[string]interface{}, 0)
		formValue := ctx.FormValue("url")
		if formValue == "" {
			data["form_result"] = "You need to enter a URL"
		} else {
			key, err := factory.Gen(formValue)
			if err != nil {
				data["form_result"] = "Invalid URL."
			} else {
				if err = db.Set(key, formValue); err != nil {
					data["form_result"] = "Internal error while saving the url"
					app.Log(iris.DevMode, "while saving url: "+err.Error())
				} else {
					ctx.SetStatusCode(iris.StatusOK)
					shortenURL := "http://" + app.Config.VHost + "/u/" + key
					data["form_result"] = template.HTML("<pre><a target='_new' href='" + shortenURL + "'/></pre>")
				}
			}
		}

		data["url_count"] = db.Len()
		ctx.Render("index.html", data)
	})

	app.Listen("localhost:8080")
}
