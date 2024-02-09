package main

import (
	"os"
	"text/template"

	"github.com/luqus/templater/types"
)

func main() {

	tmpl := "Hello {{.Name}}"

	user := types.User{"Luqus"}
	t, err := template.New("test").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
}
