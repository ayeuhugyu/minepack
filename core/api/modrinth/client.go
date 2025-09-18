package modrinth

import (
	"net/http"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
)

var ModrinthClient = modrinth.NewClient(&http.Client{})
