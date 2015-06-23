package x

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
)

var log = Log("x")

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Instruction struct {
	SubjectId   string `json:"subject_id,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`
	Predicate   string `json:"predicate,omitempty"`
	Object      []byte `json:"object,omitempty"`
	ObjectId    string `json:"object_id,omitempty"`
	NanoTs      int64  `json:"nano_ts,omitempty"`
	Source      string `json:"source,omitempty"`
}

type Its []Instruction

func (its Its) Len() int           { return len(its) }
func (its Its) Swap(i, j int)      { its[i], its[j] = its[j], its[i] }
func (its Its) Less(i, j int) bool { return its[i].NanoTs < its[j].NanoTs }

const (
	E_ERROR            = "E_ERROR"
	E_INVALID_METHOD   = "E_INVALID_METHOD"
	E_INVALID_REQUEST  = "E_INVALID_REQUEST"
	E_INVALID_USER     = "E_INVALID_USER"
	E_MISSING_REQUIRED = "E_MISSING_REQUIRED"
	E_OK               = "E_OK"
	E_UNAUTHORIZED     = "E_UNAUTHORIZED"
)

func Log(p string) *logrus.Entry {
	logrus.SetLevel(logrus.DebugLevel)
	l := logrus.WithFields(logrus.Fields{
		"package": p,
	})
	return l
}

func LogErr(entry *logrus.Entry, err error) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"error": err.Error(),
	})
}

func SetStatus(w http.ResponseWriter, code, msg string) {
	r := &Status{Code: code, Message: msg}
	if js, err := json.Marshal(r); err == nil {
		fmt.Fprint(w, string(js))
	} else {
		panic(fmt.Sprintf("Unable to marshal: %+v", r))
	}
}

func ParseRequest(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		SetStatus(w, E_ERROR, fmt.Sprintf("While parsing request: %v", err))
		return false
	}
	return true
}

const alphachars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func UniqueString(alpha int) string {
	var buf bytes.Buffer
	for i := 0; i < alpha; i++ {
		idx := rand.Intn(len(alphachars))
		buf.WriteByte(alphachars[idx])
	}
	return buf.String()
}

func ParseIdFromUrl(r *http.Request, urlToken string) (uid string, ok bool) {
	url := r.URL.Path
	idx := strings.LastIndex(url, urlToken)
	if idx == -1 {
		return
	}
	return url[idx+len(urlToken):], true
}

func Reply(w http.ResponseWriter, rep interface{}) {
	if js, err := json.Marshal(rep); err == nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(js))
	} else {
		SetStatus(w, E_ERROR, "Internal server error")
	}
}
