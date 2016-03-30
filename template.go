package main

import "html/template"

type tmpl struct {
	*template.Template
}

type tmplData struct {
	Active  int
	Host    string
	Data    map[string]interface{}
	Player  *userInfo
	Scripts []template.URL
}

func (t *tmpl) ExecuteTemplate(c context, name string, data tmplData) error {
	data.Host = c.r.Host
	data.Player = c.u

	if err := t.Template.ExecuteTemplate(c.w, "head.html", data); err != nil {
		return err
	}

	if err := t.Template.ExecuteTemplate(c.w, name, data); err != nil {
		return err
	}

	if err := t.Template.ExecuteTemplate(c.w, "foot.html", data); err != nil {
		return err
	}

	return nil
}
