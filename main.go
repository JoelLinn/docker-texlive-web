package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// run a simple http server to receive tex files and return compiled file
func main() {
	staticdir := "/var/www/html"

	tex_timeout := func() time.Duration {
		tex_timeout_ms, err := strconv.ParseUint(os.Getenv("TEXLIVE_WEB_TEX_TIMEOUT_MS"), 10, 32)
		_ = err
		return time.Duration(tex_timeout_ms) * time.Millisecond
	}()

	// serve a simple form to upload a .tex file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving index.html")
		http.ServeFile(w, r, filepath.Join(staticdir, "index.html"))
	})

	// try to compile the tex file
	http.HandleFunc("/pdflatex", func(w http.ResponseWriter, r *http.Request) {
		var texh io.Reader // handler to uploaded tex file

		ctype, exist := r.Header["Content-Type"]
		if exist && (ctype[0] == "application/x-tex" || ctype[0] == "text/x-tex") {
			texh = r.Body
		} else {
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
			texh, err = texs[0].Open()
			if err != nil {
				log.Println("Error opening file. " + err.Error())
				w.WriteHeader(500)
				return
			}
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
		if _, err := io.Copy(tmpf, texh); err != nil {
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
			log.Println("Starting compilation.")
			if err := pdflatex.Start(); err != nil {
				log.Println("Error starting pdflatex: " + err.Error())
				w.WriteHeader(500)
				return
			}

			var timer *time.Timer = nil
			if tex_timeout > 0 {
				timer = time.AfterFunc(tex_timeout, func() {
					log.Println("Compilation timeout.")
					pdflatex.Process.Kill()
				})
			}
			err := pdflatex.Wait()
			if timer != nil {
				timer.Stop()
			}
			if err != nil {
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

	log.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
