package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRewriteRequestURL_AppendsPathAndQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/dir?b=1", nil)
	target := &url.URL{Scheme: "http", Host: "example.com:8080", Path: "/base", RawQuery: "a=1"}

	RewriteRequestURL(req, target)

	if req.URL.Scheme != "http" || req.URL.Host != "example.com:8080" {
		t.Fatalf("scheme/host not set: %v", req.URL)
	}
	if req.URL.Path != "/base/dir" {
		t.Fatalf("expected /base/dir got %s", req.URL.Path)
	}
	// RawQuery can be either order
	if req.URL.RawQuery != "a=1&b=1" && req.URL.RawQuery != "b=1&a=1" {
		t.Fatalf("unexpected raw query: %s", req.URL.RawQuery)
	}
}

func TestJoinURLPath_RawPaths(t *testing.T) {
	a := &url.URL{Path: "/a/", RawPath: "/a/"}
	b := &url.URL{Path: "b", RawPath: "b"}
	p, rp := JoinURLPath(a, b)
	if p != "/a/b" {
		t.Fatalf("unexpected path: %s", p)
	}
	if rp != "/a/b" {
		t.Fatalf("unexpected rawpath: %s", rp)
	}
}

func TestSingleJoiningSlash(t *testing.T) {
	if SingleJoiningSlash("/a/", "/b") != "/a/b" {
		t.Fatalf("unexpected join")
	}
	if SingleJoiningSlash("/a", "b") != "/a/b" {
		t.Fatalf("unexpected join")
	}
}
