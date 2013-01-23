package main

import "html/template"
import "net/url"
import "path/filepath"

var funcs = template.FuncMap{
  "SolutionFor": SolutionFor,
  "NumSolved": NumSolved,
  "FormGet": FormGet,
}

func FormGet(form url.Values, key string) string {
  val := form[key]
  if val != nil && len(val) > 0 {
    return val[0]
  }
  return ""
}

func AdminTemplate(names... string) *template.Template {
  return Template("_admin.html", names...)
}

func Template(layout string, names... string) *template.Template {
  t := template.New(layout).Funcs(funcs)
  paths := make([]string, len(names) + 1)
  paths[0] = filepath.Join("templates", layout)
  for i, name := range names {
    paths[i + 1] = filepath.Join("templates", name)
  }

  return template.Must(t.ParseFiles(paths...))
}
