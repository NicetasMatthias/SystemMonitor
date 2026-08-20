package web

import "embed"

//go:embed templates/*.html
//go:embed static/css/*
//go:embed static/js/*

var FS embed.FS
