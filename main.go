package main

import (
  "net/http"
  // "io"
  "html/template"
  "encoding/json"
  // "log"
  // "fmt"
)

type User struct {
  Username string `json:"username"`
  Age int `json:"age"`
  IsAdmin bool `json:"is_admin"`
}

type Template struct {
  templates *template.Template
}

// create a context struct that will be passed to the route handlers
// that contains the http.ResponseWriter, *http.Request and *Template
type Context struct {
  w http.ResponseWriter
  req *http.Request
  t *Template
}

// create a render function (on context) that takes a filename and data and executes the template
func (c *Context) render(fileName string, data any) error {
  return c.t.templates.ExecuteTemplate(c.w, fileName, data)
}

// create a writeJSON function (on context) that takes a status code and data and encodes it as JSON
func (c *Context) writeJSON(code int, v any) error {
  c.w.Header().Set("Content-Type", "application/json")
  c.w.WriteHeader(code)
  return json.NewEncoder(c.w).Encode(v)
}

func main() { 
  // create a new template
  t := &Template{
    templates: template.Must(template.ParseGlob("www/*.html")),
  }

  // create routes and listen on port 8080
  http.HandleFunc("/", makeAPIHandler(handleHome, t))
  http.HandleFunc("/api/user", makeAPIHandler(handleUsers, t))
  http.HandleFunc("/cat/facts", makeAPIHandler(showCatFacts, t))

  http.ListenAndServe(":8080", nil)
}

// use type to represent a function that takes a http.ResponseWriter and a *http.Request and returns an error
type apiFunc func(c *Context) error

// create function that takes an apiFunc and returns a http.HandlerFunc (used in the http.HandleFunc)
func makeAPIHandler(fn apiFunc, t *Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    ctx := &Context{
      w: w,
      req: r,
      t: t,
    }
    err := fn(ctx)
    if err != nil {
      // handle error
      ctx.writeJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
  }
}

// route handlers
func handleUsers(ctx *Context) error {
  return ctx.writeJSON(http.StatusOK, map[string]string{"name": "User"})
}

func handleHome(ctx *Context) error {
  user := &User{
    Username: "User",
    Age: 20,
    IsAdmin: false,
  }
  return ctx.render("index.html", user)
}

func showCatFacts(ctx *Context) error {
  facts, err := fetchCatFacts()
  if err != nil {
    return err
  }
  return ctx.render("facts.html", facts)
}

type CatFact struct {
  Fact string `json:"fact"`
}

type CFResponse struct {
  Data []CatFact `json:"data"`
}

func fetchCatFacts() ([]CatFact, error) {
  res, err := http.Get("https://catfact.ninja/facts")
  if err != nil {
    return nil, err
  }

  defer res.Body.Close()

  var catFacts CFResponse

  err = json.NewDecoder(res.Body).Decode(&catFacts)
  if err != nil {
    return nil, err
  }

  return catFacts.Data, nil
}
