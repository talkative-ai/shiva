package prospectacle

import (
	"io"
	"net/http"
)

func init() {
	http.HandleFunc("/.well-known/acme-challenge/NX1lFw-WUrXPNfjUPvgEmzulsrpsobWGa3SW_GO5MPQ", root)
}

// [START func_root]
func root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "NX1lFw-WUrXPNfjUPvgEmzulsrpsobWGa3SW_GO5MPQ.dHP9qyK89IfuTqoierLcfa3E_gZetc7DP7B5SuAznMk")
}

// [END func_root]
