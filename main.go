package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/tritonmedia/identifier/pkg/rabbitmq"
	api "github.com/tritonmedia/tritonmedia.go/pkg/proto"
	"github.com/tritonmedia/twilight.go/pkg/parser"
	"github.com/tritonmedia/twilight.go/pkg/storage"
	"github.com/tritonmedia/twilight.go/pkg/storage/fs"
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

func reciever(s storage.Provider, rabbit *rabbitmq.Client, w http.ResponseWriter, r *http.Request) {
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

		mediaID := r.Header.Get("X-Media-ID")
		if mediaName == "" {
			log.Errorf("missing X-Media-ID")
			sendError(w, http.StatusBadRequest, "missing X-Media-ID")
			return
		}

		mediaQuality := r.Header.Get("X-Media-Quality")
		if mediaQuality == "" {
			log.Errorf("missing X-Media-Quality")
			sendError(w, http.StatusBadRequest, "missing X-Media-Quality")
			return
		}

		var itypeID int32
		var ok bool
		if itypeID, ok = api.Media_MediaType_value[strings.ToUpper(mediaType)]; !ok {
			sendError(w, http.StatusBadRequest, "invalid media type")
			return
		}

		typeID := api.Media_MediaType(itypeID)

		newName := fmt.Sprintf("%s.mkv", mediaName)

		// if movie, OK to leave m empty because we do type detection on the other end
		m := parser.Metadata{}
		if mediaType != "movie" {
			m, err = parser.ParseFile(p.FileName())
			if err != nil {
				log.Errorf("failed to parse file: %v", err)
				sendError(w, http.StatusInternalServerError, "failed to parse filename")
				return
			}

			newName = fmt.Sprintf("%s - S%dE%d.mkv", mediaName, m.Season, m.Episode)
		}

		key := fmt.Sprintf("%s/%s/%s", strings.ToLower(typeID.String()), mediaName, newName)

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

		log.Infof("uploaded file to remote")

		log.Infof("creating v1.identify.newfile message")
		i := api.IdentifyNewFile{
			CreatedAt: time.Now().Format(time.RFC3339),
			Quality:   mediaQuality,
			Key:       key,
			Episode:   int64(m.Episode),
			Season:    int64(m.Season),
			Media: &api.Media{
				Id:   mediaID,
				Type: typeID,
			},
		}

		b, err := proto.Marshal(&i)
		if err != nil {
			panic(err)
		}
		if err := rabbit.Publish("v1.identify.newfile", b); err != nil {
			log.Errorf("failed to create message: %v", err)
			sendError(w, http.StatusInternalServerError, "failed to publish message")
		}
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

	provider := strings.ToLower(os.Getenv("TWILIGHT_STORAGE_PROVIDER"))
	if provider == "" { // default to the fs client
		provider = "fs"
	}

	log.Infof("creating storage client (%s)...", provider)

	var s storage.Provider
	switch provider {
	case "s3":
		// TODO(jaredallard): add support for other clients
		var err error
		s, err = s3.NewProvider(
			os.Getenv("TWILIGHT_S3_ACCESS_KEY"),
			os.Getenv("TWILIGHT_S3_SECRET_KEY"),
			os.Getenv("TWILIGHT_S3_ENDPOINT"),
			os.Getenv("TWILIGHT_S3_BUCKET"),
		)
		if err != nil {
			log.Fatalf("failed to create s3 client: %v", err)
		}
		break
	case "fs":
		s = fs.NewProvider(os.Getenv("TWILIGHT_FS_BASE"))
		break
	default:
		log.Fatalf("invalid storage provider '%s'", provider)
	}

	amqpEndpoint := os.Getenv("TWILIGHT_RABBITMQ_ENDPOINT")
	if amqpEndpoint == "" {
		amqpEndpoint = "amqp://user:bitnami@127.0.0.1:5672"
		log.Warnf("TWILIGHT_RABBITMQ_ENDPOINT not defined, defaulting to local config: %s", amqpEndpoint)
	}

	client, err := rabbitmq.NewClient(amqpEndpoint)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	port := ":3402"
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	}

	http.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		reciever(s, client, w, r)
	})

	log.Infof("listening on port %s", port)
	log.Fatalf("Exited: %s", http.ListenAndServe(port, nil))
}
