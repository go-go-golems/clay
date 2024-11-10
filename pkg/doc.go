package pkg

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed filewalker/doc/*
//go:embed filefilter/doc/*
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
