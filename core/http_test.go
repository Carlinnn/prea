package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prea-framework/prea/logger"

	"github.com/gin-gonic/gin"
)

// events records what ran, in order, for the request under test.
var (
	flowMu  sync.Mutex
	flowLog []string
)

func record(e string) { flowMu.Lock(); flowLog = append(flowLog, e); flowMu.Unlock() }

func takeEvents() []string {
	flowMu.Lock()
	defer flowMu.Unlock()
	got := flowLog
	flowLog = nil
	return got
}

// layer records itself before and after the rest of the chain.
func layer(name string) Middleware {
	return func(c *gin.Context) {
		record(name + ">")
		c.Next()
		record("<" + name)
	}
}

// counter is a middleware component, created and injected by the container.
type counter struct{}

func (*counter) Handle(c *gin.Context) {
	record("component>")
	c.Next()
	record("<component")
}

type flowRoutes struct{}

func (flowRoutes) Register(r gin.IRouter) {
	g := r.Group("/flow")
	g.GET("/items", layer("route"), func(c *gin.Context) {
		record("handler")
		c.Header("X-Handler", "items")
		Reply(c, http.StatusOK, Data([]string{"a", "b"}))
	})
	g.GET("/custom", func(c *gin.Context) { c.Header("X-Which", "get"); c.String(200, "get") })
	g.HEAD("/custom", func(c *gin.Context) { c.Header("X-Which", "head"); c.Status(200) })
	g.GET("/panic", func(c *gin.Context) { record("handler"); panic("boom") })
	g.GET("/error", func(c *gin.Context) {
		Error(c, errors.Join(errors.New("db said no"), NewError(http.StatusForbidden, "not yours")))
	})
	g.POST("/write", func(c *gin.Context) { record("handler"); c.Status(http.StatusCreated) })
	g.GET("/abort-route", func(c *gin.Context) { record("route-abort"); c.AbortWithStatus(http.StatusUnauthorized) },
		func(c *gin.Context) { record("handler") })
}

type gateRoutes struct{}

func (gateRoutes) Register(r gin.IRouter) {
	r.GET("/gate/x", func(c *gin.Context) { record("handler") })
}

func flowApp(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	prev := logger.Log
	logger.Log = logger.New(io.Discard, slog.LevelError, false) // the panic test logs a stack
	t.Cleanup(func() { logger.Log = prev })
	gate := func(c *gin.Context) { record("gate-abort"); c.AbortWithStatus(http.StatusForbidden) }
	engine, closer, err := App{
		Modules: []*Module{
			{Name: "flow", Providers: []any{flowRoutes{}}, Middlewares: []any{layer("module")}},
			{Name: "gate", Providers: []any{gateRoutes{}}, Middlewares: []any{Middleware(gate)}},
		},
		Security: Security{
			Headers:     map[string]string{"X-Secure": "yes"},
			CORS:        &CORS{AllowOrigins: []string{"https://app.example"}},
			CSRF:        &CSRF{},
			Middlewares: []any{layer("security")},
		},
		Middlewares: []any{layer("global"), Recovery, counter{}},
	}.Build(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closer(context.Background()) })
	takeEvents()
	return engine
}

func serveFlow(engine http.Handler, method, path string, header ...string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	for i := 0; i+1 < len(header); i += 2 {
		r.Header.Set(header[i], header[i+1])
	}
	engine.ServeHTTP(w, r)
	return w
}

func TestHTTPFlow(t *testing.T) {
	engine := flowApp(t)
	full := []string{"security>", "global>", "component>", "module>", "route>", "handler", "<route", "<module", "<component", "<global", "<security"}

	t.Run("every layer runs once, in order, and unwinds in reverse", func(t *testing.T) {
		w := serveFlow(engine, "GET", "/flow/items")
		if got := takeEvents(); !slices.Equal(got, full) {
			t.Fatalf("got  %v\nwant %v", got, full)
		}
		if w.Code != 200 || w.Header().Get("X-Secure") != "yes" || w.Header().Get("X-Handler") != "items" || !strings.Contains(w.Body.String(), `"a"`) {
			t.Fatalf("%d %v %s", w.Code, w.Header(), w.Body)
		}
	})

	t.Run("HEAD runs the GET route once, with the same headers", func(t *testing.T) {
		get := serveFlow(engine, "GET", "/flow/items")
		takeEvents()
		head := serveFlow(engine, "HEAD", "/flow/items")
		if got := takeEvents(); !slices.Equal(got, full) {
			t.Fatalf("HEAD must run the chain exactly once:\ngot  %v\nwant %v", got, full)
		}
		if head.Code != 200 {
			t.Fatalf("HEAD: %d", head.Code)
		}
		for _, h := range []string{"X-Secure", "X-Handler", "Content-Type"} {
			if head.Header().Get(h) != get.Header().Get(h) {
				t.Errorf("%s: HEAD %q, GET %q", h, head.Header().Get(h), get.Header().Get(h))
			}
		}
		// CORS adds "Vary: Origin": security running twice would repeat it.
		if !slices.Equal(head.Header().Values("Vary"), get.Header().Values("Vary")) {
			t.Errorf("security ran twice: Vary %v, GET %v", head.Header().Values("Vary"), get.Header().Values("Vary"))
		}
	})

	t.Run("a HEAD route of the app wins", func(t *testing.T) {
		if w := serveFlow(engine, "HEAD", "/flow/custom"); w.Header().Get("X-Which") != "head" {
			t.Fatalf("got %q", w.Header().Get("X-Which"))
		}
		if w := serveFlow(engine, "GET", "/flow/custom"); w.Header().Get("X-Which") != "get" {
			t.Fatalf("got %q", w.Header().Get("X-Which"))
		}
	})

	t.Run("404 goes through security and the global middlewares, once", func(t *testing.T) {
		for _, method := range []string{"GET", "HEAD"} {
			takeEvents()
			w := serveFlow(engine, method, "/nowhere")
			want := []string{"security>", "global>", "component>", "<component", "<global", "<security"}
			if got := takeEvents(); w.Code != 404 || w.Header().Get("X-Secure") != "yes" || !slices.Equal(got, want) {
				t.Fatalf("%s: %d %q %v", method, w.Code, w.Header().Get("X-Secure"), got)
			}
		}
	})

	takeEvents()
	t.Run("abort in a preá middleware stops the rest, the outer layers unwind", func(t *testing.T) {
		w := serveFlow(engine, "GET", "/gate/x")
		want := []string{"security>", "global>", "component>", "gate-abort", "<component", "<global", "<security"}
		if got := takeEvents(); w.Code != 403 || !slices.Equal(got, want) {
			t.Fatalf("%d %v", w.Code, got)
		}
	})

	t.Run("abort in a route middleware skips the handler", func(t *testing.T) {
		w := serveFlow(engine, "GET", "/flow/abort-route")
		if got := takeEvents(); w.Code != 401 || slices.Contains(got, "handler") || !slices.Contains(got, "<module") {
			t.Fatalf("%d %v", w.Code, got)
		}
	})

	t.Run("a panic becomes a 500, the outer layers unwind", func(t *testing.T) {
		w := serveFlow(engine, "GET", "/flow/panic")
		got := takeEvents()
		if w.Code != 500 || !strings.Contains(w.Body.String(), "internal server error") {
			t.Fatalf("%d %s", w.Code, w.Body)
		}
		if !slices.Equal(got[len(got)-2:], []string{"<global", "<security"}) {
			t.Fatalf("the layers outside Recovery must finish: %v", got)
		}
	})

	t.Run("CORS preflight is answered by security alone", func(t *testing.T) {
		w := serveFlow(engine, "OPTIONS", "/flow/write", "Origin", "https://app.example", "Access-Control-Request-Method", "POST")
		if got := takeEvents(); w.Code != 204 || len(got) != 0 || w.Header().Get("Access-Control-Allow-Origin") != "https://app.example" {
			t.Fatalf("%d %v %v", w.Code, w.Header(), got)
		}
	})

	t.Run("CSRF blocks a cross-site write before any middleware", func(t *testing.T) {
		w := serveFlow(engine, "POST", "/flow/write", "Origin", "https://evil.example", "Sec-Fetch-Site", "cross-site")
		if got := takeEvents(); w.Code != 403 || len(got) != 0 {
			t.Fatalf("%d %v", w.Code, got)
		}
		w = serveFlow(engine, "POST", "/flow/write", "Origin", "https://app.example", "Sec-Fetch-Site", "cross-site")
		if got := takeEvents(); w.Code != 201 || !slices.Contains(got, "handler") {
			t.Fatalf("a CORS origin is trusted: %d %v", w.Code, got)
		}
	})

	t.Run("Error replies with the status in the chain", func(t *testing.T) {
		w := serveFlow(engine, "GET", "/flow/error")
		if w.Code != 403 || !strings.Contains(w.Body.String(), `"text":"not yours"`) || strings.Contains(w.Body.String(), "db said no") {
			t.Fatalf("%d %s", w.Code, w.Body)
		}
	})
}

// Over a real connection, HEAD gets the GET's headers and no body, and the
// request log names the method the client sent.
func TestHEADOverTheWire(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	prev := logger.Log
	logger.Log = logger.New(&logs, slog.LevelDebug, false)
	t.Cleanup(func() { logger.Log = prev })

	app := App{
		Modules:     []*Module{{Name: "flow", Providers: []any{flowRoutes{}}}},
		Middlewares: []any{RequestLogger},
		Server:      Server{ShutdownTimeout: time.Second},
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Serve(ctx, ln) }()
	t.Cleanup(func() { stop(); <-done })
	url := "http://" + ln.Addr().String() + "/flow/items"

	var get *http.Response
	for range 50 {
		if get, err = http.Get(url); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(get.Body)
	get.Body.Close()

	head, err := http.Head(url)
	if err != nil {
		t.Fatal(err)
	}
	headBody, _ := io.ReadAll(head.Body)
	head.Body.Close()
	if head.StatusCode != 200 || len(headBody) != 0 {
		t.Fatalf("HEAD: %d with %d bytes of body", head.StatusCode, len(headBody))
	}
	if head.ContentLength != int64(len(body)) || head.Header.Get("X-Handler") != "items" {
		t.Fatalf("HEAD must have the GET's headers: length %d (GET body %d), %v", head.ContentLength, len(body), head.Header)
	}
	if !strings.Contains(logs.String(), "method=HEAD path=/flow/items status=200") {
		t.Fatalf("the request log must name HEAD:\n%s", logs.String())
	}
}
