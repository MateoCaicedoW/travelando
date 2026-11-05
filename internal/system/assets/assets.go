package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed *
var Files embed.FS

// IsDev reports whether assets should come from the Vite dev server.
func IsDev() bool {
	return os.Getenv("GO_ENV") != "production"
}

// Path returns a Vite URL in development and an embedded bundle URL in production.
func Path(path string) string {
	if IsDev() {
		return "http://localhost:5173" + path
	}
	return "/assets/dist" + path
}

type fileHandler struct {
	fs          embed.FS
	stripPrefix string
	fileServer  http.Handler
}

func (fh *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := path.Clean(strings.TrimPrefix(r.URL.Path, fh.stripPrefix))
	if name == "." {
		http.NotFound(w, r)
		return
	}

	stat, err := fs.Stat(fh.fs, name)
	if err == nil && stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	fh.fileServer.ServeHTTP(w, r)
}

// NewAssetHandler serves embedded files without exposing directory listings.
func NewAssetHandler(fsys embed.FS, stripPrefix string) http.Handler {
	return &fileHandler{
		fs:          fsys,
		stripPrefix: stripPrefix,
		fileServer:  http.StripPrefix(stripPrefix, http.FileServer(http.FS(fsys))),
	}
}
