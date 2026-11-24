// Package web holds the browser test page that every target serves.
package web

import "embed"

// Files is the page, its script and its styles.
//
//go:embed index.html app.js style.css
var Files embed.FS
