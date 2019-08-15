package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	humanize "github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
	"github.com/tritonmedia/twilight.go/pkg/parser"
	"github.com/tritonmedia/twilight.go/pkg/storage"
	"github.com/tritonmedia/twilight.go/pkg/storage/s3"
)

// HTTPError is an error returned by Twilight
type HTTPError struct {
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// HTTPSuccess is sent on successful upload
type HTTPSuccess struct {
	Message string `json:"message"`
}

// sendError sends a standard HTTPError to the client
func sendError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	e := HTTPError{
		Message:   msg,
		Retryable: false,
	}

	b, err := json.Marshal(e)
	if err != nil {
		return
	}

	w.WriteHeader(statusCode)
	w.Write(b)
}

func reciever(s storage.Provider, w http.ResponseWriter, r *http.Request) {
	// all requests use json anywa
	w.Header().Set("Content-Type", "application/json")

	// old endpoint
	if r.Method == "POST" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"unsupported enpoint, please use just PUT /v1/media"}`))
		return
	}

	cl, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		log.Warnf("failed to read content-length, setting to 0")
		cl = 0
	}

	log.Infof("processing media upload, length='%s'", humanize.Bytes(uint64(cl)))

	mr, err := r.MultipartReader()
	if err != nil {
		log.Errorf("Hit error while opening multipart reader: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to parse multipart data")
		return
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF { // only hit if we didn't get a file
			log.Warnf("hit EOF: %v", err)
			sendError(w, http.StatusInternalServerError, "missing file")
			return
		} else if err != nil {
			log.Errorf("failed to read file: %v", err)
			sendError(w, http.StatusInternalServerError, "failed to parse multipart data after read")
			return
		}

		if p.FormName() != "file" {
			log.Infof("skipping unknown field: %v", p.FormName())
			continue
		}

		log.Infof("uploading file ...")

		mediaType := r.Header.Get("X-Media-Type")
		if mediaType == "" {
			log.Errorf("missing X-Media-Type")
			sendError(w, http.StatusBadRequest, "missing X-Media-Type")
			return
		}

		mediaName := r.Header.Get("X-Media-Name")
		if mediaName == "" {
			log.Errorf("missing X-Media-Name")
			sendError(w, http.StatusBadRequest, "missing X-Media-Name")
			return
		}

		newName := fmt.Sprintf("%s.mkv", mediaName)
		if mediaType != "movie" {
			m, err := parser.ParseFile(p.FileName())
			if err != nil {
				log.Errorf("failed to parse file: %v", err)
				sendError(w, http.StatusInternalServerError, "failed to parse filename")
				return
			}

			newName = fmt.Sprintf("%s - S%dE%d.mkv", mediaName, m.Season, m.Episode)
		}

		// TODO(jaredallard): better key calculation
		key := fmt.Sprintf("%s/%s/%s", mediaType, mediaName, newName)

		log.Infof("uploading file to '%s'", key)

		err = s.Create(p, key)
		if err == storage.ErrorIsExists {
			sendError(w, http.StatusConflict, "file already exists")
			return
		} else if err != nil {
			log.Errorf("failed to write file to storage provider: %v", err)
			sendError(w, http.StatusInternalServerError, "failed to stream to storageprovider")
			return
		}

		// asssuming done, so break out
		log.Infof("uploaded file to remote")
		break
	}

	b, err := json.Marshal(HTTPSuccess{
		Message: "Succesfully uploaded file.",
	})
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to send success response")
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		log.Errorf("failed to write success to client: %v", err)
	}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	if os.Getenv("TWILIGHT_DEBUG") != "" {
		log.SetReportCaller(true)
	}

	log.Infof("creating storage client ...")
	// TODO(jaredallard): add support for other clients
	s, err := s3.NewProvider(
		os.Getenv("TWILIGHT_S3_ACCESS_KEY"),
		os.Getenv("TWILIGHT_S3_SECRET_KEY"),
		os.Getenv("TWILIGHT_S3_ENDPOINT"),
		os.Getenv("TWILIGHT_S3_BUCKET"),
	)
	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}

	port := ":3402"
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	}

	http.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		reciever(s, w, r)
	})

	log.Infof("listening on port %s", port)
	log.Fatalf("Exited: %s", http.ListenAndServe(port, nil))
}
