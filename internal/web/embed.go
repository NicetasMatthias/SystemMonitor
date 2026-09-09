package web

import "embed"

//go:embed templates/*.html
//go:embed static/css/*
//go:embed static/js/*
//go:embed static/favicon.svg

var FS embed.FS
