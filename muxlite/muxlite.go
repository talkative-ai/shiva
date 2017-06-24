package muxlite

import (
	"net/http"
	"strings"
)

type Method string

const (
	GET  Method = "GET"
	POST        = "POST"
	ALL         = "ALL"
)

type PathNode struct {
	Value    string
	Children map[string]*PathNode
	Routes   map[Method]*Route
}

type Router struct {
	routes map[string]*PathNode
}

type Route struct {
	path       string
	handler    http.HandlerFunc
	methods    []string
	parentNode *PathNode
}

func trim(s string, char byte) string {
	for s[0] == char {
		s = s[1:]
	}
	for s[len(s)-1] == char {
		s = s[:len(s)-1]
	}

	return s
}

func prune(s string, char byte) string {
	var prev byte
	for i := 0; i < len(s); i++ {
		if s[i] == prev && s[i] == char {
			s = s[:i] + s[i+1:]
			i--
		}
		prev = s[i]
	}
	return s
}

func (r *Route) Methods(methods ...Method) {
	for routeMethod, route := range r.parentNode.Routes {
		if route == r {
			r.parentNode.Routes[routeMethod] = nil
		}
	}
	for _, method := range methods {
		r.parentNode.Routes[method] = r
	}
}

func (r *Router) Handle(path string, handler http.HandlerFunc) *Route {
	newRoute := &Route{
		path:    path,
		handler: handler,
		methods: []string{},
	}

	path = trim(prune(path, '/'), '/')
	slugs := strings.Split(path, "/")

	var currentNode *PathNode
	currentNodeMap := r.routes
	for _, slug := range slugs {
		if _, exist := currentNodeMap[slug]; !exist {
			newNode := &PathNode{
				Value:    slugs[0],
				Children: map[string]*PathNode{},
			}
			currentNodeMap[slug] = newNode
		}

		currentNode = currentNodeMap[slug]
		currentNodeMap = currentNodeMap[slug].Children
	}

	newRoute.parentNode = currentNode

	if currentNode.Routes == nil {
		currentNode.Routes = map[Method]*Route{}
	}

	currentNode.Routes[ALL] = newRoute

	return newRoute
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := trim(prune(req.URL.Path, '/'), '/')
	slugs := strings.Split(path, "/")
	var currentNode *PathNode
	currentNodeMap := r.routes
	fail := false

	for _, slug := range slugs {
		if _, exist := currentNodeMap[slug]; !exist {
			fail = true
			break
		}
		currentNode = currentNodeMap[slug]
		currentNodeMap = currentNodeMap[slug].Children
	}

	if !fail {
		for _, route := range currentNode.Routes {
			for _, method := range route.methods {
				if method == req.Method {
					route.handler(w, req)
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
	return
}

func NewRouter() *Router {
	return &Router{
		routes: map[string]*PathNode{},
	}
}
