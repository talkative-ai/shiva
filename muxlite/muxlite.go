package muxlite

import (
	"net/http"
	"strings"
)

type PathNode struct {
	Value    string
	Children map[string]*PathNode
	Routes   []*Route
}

type Router struct {
	routes map[string]*PathNode
}

type Route struct {
	path    string
	handler http.HandlerFunc
	methods []string
	router  *Router
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

func (r *Route) Methods(methods ...string) {
	for _, method := range methods {
		r.methods = append(r.methods, method)
	}
}

func (r *Router) Handle(path string, handler http.HandlerFunc) *Route {
	newRoute := &Route{
		path:    path,
		handler: handler,
		methods: []string{},
		router:  r,
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

	if currentNode.Routes == nil {
		currentNode.Routes = []*Route{}
	}

	currentNode.Routes = append(currentNode.Routes, newRoute)

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
