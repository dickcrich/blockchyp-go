// +build tools

package tools

import (
	// Build and CI/CD tools
	_ "github.com/golang/lint/golint"
	_ "github.com/josephspurrier/goversioninfo"
	_ "github.com/jstemmer/go-junit-report"
	_ "github.com/mgechev/revive"
)
