package main

import "github.com/gorilla/mux"
import "net/http"
import . "github.com/FMNSSun/plainsrv/format"
import "os"
import "log"
import "path/filepath"
import "io/ioutil"
import "io"
import "html"
import "net/url"
import "strings"

type ApiEnv struct {
	Namespaces map[string]NamespaceConfig
}

type NamespaceConfig struct {
	BasePath     string
	ContentTypes map[string]string
}

func (e *ApiEnv) ContentType(prefix string, fname string) string {
	nsconf, ok := e.Namespaces[prefix]

	if !ok {
		return "text/plain"
	}

	if nsconf.ContentTypes == nil {
		return "application/octet-stream"
	}

	ct, ok := nsconf.ContentTypes[filepath.Ext(fname)]

	if !ok {
		return "application/octet-stream"
	}

	return ct
}

func (e *ApiEnv) BasePath(prefix string) (string, bool) {
	nsconf, ok := e.Namespaces[prefix]

	if !ok {
		return prefix, false
	}

	return nsconf.BasePath, true
}

func openWithIndex(basePath string, relPath string) (string, *os.File, error) {
	file, err := os.OpenFile(filepath.Join(basePath, relPath), os.O_RDONLY, 0)

	if err != nil {
		return relPath, nil, err
	}

	fi, err := file.Stat()

	if err != nil {
		return relPath, nil, err
	}

	if fi.IsDir() {
		return openWithIndex(basePath, filepath.Join(relPath, "Home"))
	}

	return relPath, file, nil
}

func sendErr(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	io.WriteString(w, "<b style=\"color: red;\">ERROR</b></main</body></html>")
}

func returnNotFound(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "NOT FOUND")
}

func returnInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, "INTERNAL SERVER ERROR")
}

func getPath(s string) string {
	parts := strings.Split(s, "/")

	for i := 0; i < len(parts); i++ {
		if parts[i] == "." || parts[i] == ".." {
			parts[i] = ""
		}
	}

	return filepath.Join(parts...)
}

func (e *ApiEnv) rawGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	vars := mux.Vars(r)
	requestPath := getPath(vars["it"])

	basePath, ok := e.BasePath(vars["prefix"])

	if !ok {
		returnNotFound(w, nil)
		return
	}

	_, file, err := openWithIndex(basePath, requestPath)

	if os.IsNotExist(err) {
		returnNotFound(w, err)
		return
	}

	if err != nil {
		returnInternalError(w, err)
		return
	}

	io.Copy(w, file)
}

func (e *ApiEnv) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	requestPath := getPath(vars["it"])

	basePath, ok := e.BasePath(vars["prefix"])

	if !ok {
		returnNotFound(w, nil)
		return
	}

	relPath, file, err := openWithIndex(basePath, requestPath)

	navOnly := false

	if err != nil {
		if !os.IsNotExist(err) {
			returnInternalError(w, err)
			return
		} else {
			navOnly = true
		}
	}

	fullPath := filepath.Join(basePath, relPath)
	fullDir := filepath.Dir(fullPath)
	relDir := filepath.Dir(relPath)

	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "<html><head><title>")
	io.WriteString(w, html.EscapeString(relPath))
	io.WriteString(w, "</title></head><body><nav><ol>")

	files, err := ioutil.ReadDir(fullDir)

	if err != nil {
		io.WriteString(w, "</ol></nav><main>")
		sendErr(w, err)
		return
	}

	if relDir != "." {
		io.WriteString(w, "<li><a href=\"./../\">..</a></li>")
	}

	for _, fname := range files {
		fpath := url.PathEscape(fname.Name())

		if fname.IsDir() {
			fpath += "/"
		}

		io.WriteString(w, "<li><a href=\"./"+html.EscapeString(fpath)+"\">"+html.EscapeString(fpath)+"</a></li>")
	}

	io.WriteString(w, "</ol></nav><main>")

	if err != nil {
		sendErr(w, err)
		return
	}

	if !navOnly {
		Format(file, w)
	}

	io.WriteString(w, "</main></body></html>")
}

func NewAPI(e *ApiEnv) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/{it:[a-zA-Z0-9_\\./]*}", e.get).Methods("GET")
	r.HandleFunc("/-raw/{it:[a-zA-Z0-9_\\./]*}", e.rawGet).Methods("GET")
	r.HandleFunc("/{prefix:~[a-zA-Z0-9_]*}/{it:[a-zA-Z0-9_]*}", e.get).Methods("GET")
	r.HandleFunc("/-raw/{prefix:~[a-zA-Z0-9_]*}/{it:[a-zA-Z0-9_\\./]*}", e.rawGet).Methods("GET")

	return r
}
