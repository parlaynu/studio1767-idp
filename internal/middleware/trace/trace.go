package trace

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func New(next http.Handler) http.Handler {
	th := traceHandler{
		next: next,
	}
	return &th
}

type traceHandler struct {
	next http.Handler
}

func (th *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Tracef("%s %s \"%s %s %s\"", r.RemoteAddr, r.Host, r.Method, r.URL, r.Proto)

	ww := wrappedRespWriter{
		writer: w,
		status: http.StatusOK,
		size:   0,
	}

	th.next.ServeHTTP(&ww, r)
	log.Infof("%s %s \"%s %s %s\" %d %d", r.RemoteAddr, r.Host, r.Method, r.URL, r.Proto, ww.status, ww.size)
}

type wrappedRespWriter struct {
	writer http.ResponseWriter
	status int
	size   uint32
}

func (w *wrappedRespWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *wrappedRespWriter) Write(b []byte) (int, error) {
	w.size += uint32(len(b))
	return w.writer.Write(b)
}

func (w *wrappedRespWriter) WriteHeader(statuscode int) {
	w.status = statuscode
	w.writer.WriteHeader(statuscode)
}

func (w *wrappedRespWriter) Flush() {
	if v, ok := w.writer.(http.Flusher); ok {
		v.Flush()
	}
}
