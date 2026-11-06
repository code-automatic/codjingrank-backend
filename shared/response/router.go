package response

import "net/http"

// Router gives generated routing errors the API envelope without rewriting handler responses.
func Router(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w = headWriter{w}
		}
		handler, pattern := mux.Handler(r)
		if pattern == "" {
			// Only the mux-generated fallback is inspected; application handlers run once below.
			fallback := &routingErrorWriter{header: make(http.Header)}
			handler.ServeHTTP(fallback, r)
			if fallback.code == http.StatusNotFound || fallback.code == http.StatusMethodNotAllowed {
				if allow := fallback.header.Get("Allow"); allow != "" {
					w.Header().Set("Allow", allow)
				}
				ErrorSimple(w, fallback.code, http.StatusText(fallback.code))
				return
			}
		}
		// ServeMux must dispatch to populate PathValue and preserve canonical redirects.
		mux.ServeHTTP(w, r)
	})
}

type routingErrorWriter struct {
	header http.Header
	code   int
}

func (w *routingErrorWriter) Header() http.Header         { return w.header }
func (w *routingErrorWriter) WriteHeader(code int)        { w.code = code }
func (w *routingErrorWriter) Write(p []byte) (int, error) { return len(p), nil }

type headWriter struct{ http.ResponseWriter }

func (w headWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w headWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
