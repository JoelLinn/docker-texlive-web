package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// run a simple http server to receive tex files and return compiled file
func main() {
	staticdir := "/var/www/"

	// serve a simple form to upload a .tex file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving index.html")
		http.ServeFile(w, r, filepath.Join(staticdir, "index.html"))
	})

	// try to compile the tex file
	http.HandleFunc("/pdflatex", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(1024 * 1024 * 512)
		if err != nil {
			log.Println("Error parsing multipart form: " + err.Error())
			w.WriteHeader(400)
			return
		}

		// was a file sent?
		texs, exist := r.MultipartForm.File["tex"]
		if !exist {
			log.Println("No file was sent.")
			w.WriteHeader(400)
			return
		}
		texh, err := texs[0].Open()
		if err != nil {
			log.Println("Error opening file. " + err.Error())
			w.WriteHeader(500)
			return
		}

		// create temp dir
		tmpdir, err := ioutil.TempDir("", "tex")
		if err != nil {
			log.Println("Could not create temp dir. " + err.Error())
			w.WriteHeader(500)
			return
		}
		defer os.RemoveAll(tmpdir) // clean up

		// write uploaded file to disk
		tmpfn := filepath.Join(tmpdir, "upload.tex")
		tmpf, err := os.OpenFile(tmpfn, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Println("Could not open temp file. " + err.Error())
			w.WriteHeader(500)
			return
		}
		if written, err := io.Copy(tmpf, texh); err != nil || written != texs[0].Size {
			log.Println("Could not save file. " + err.Error())
			w.WriteHeader(500)
			return
		}
		if err := tmpf.Close(); err != nil {
			log.Panicln("Could not close file handle. " + err.Error())
			return
		}

		// compile the document
		if err := os.Chdir(tmpdir); err != nil {
			log.Panicln("Could not change into temp dir. " + err.Error())
			return
		}
		for i := 0; i < 2; i++ { // tex needs two passes to resolve references
			pdflatex := exec.Command("pdflatex", "-interaction=nonstopmode", "upload.tex")
			if err := pdflatex.Run(); err != nil {
				log.Println("Could not compile the tex file. " + err.Error())
				w.WriteHeader(500)
				return
			}
		}
		// check if a pdf was generated
		pdffn := filepath.Join(tmpdir, "upload.pdf")
		if _, err := os.Stat(pdffn); os.IsNotExist(err) {
			log.Println("No pdf output file.")
			w.WriteHeader(500)
			return
		}

		// send pdf to client
		log.Println("Serving compiled pdf")
		http.ServeFile(w, r, pdffn)
	})
	http.ListenAndServe(":8080", nil)
}
