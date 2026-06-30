package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/syumai/workers"
	"github.com/yuin/goldmark"
)

type WebringEntry struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Gh   string `json:"gh"`
}

type BlogPost struct {
	Slug     string
	Title    string
	Date     string
	Content  string
	External string
}

//go:embed webring.json
var webringRaw []byte

//go:embed index.html
var indexHTML []byte

//go:embed blog.html
var blogHTML []byte

//go:embed post.html
var postHTML []byte

//go:embed posts/*.md
var postsFS embed.FS

var hostsToIgnore = []string{"ring.seggs.lol", "seggs.lol", "www.seggs.lol", "links.seggs.lol"}

// initial returns the uppercased first character of s (for the letter avatar).
func initial(s string) string {
	for _, r := range s {
		return strings.ToUpper(string(r))
	}
	return ""
}

// host strips the scheme and a leading "www." from a url for display.
func host(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	return strings.TrimPrefix(u.Host, "www.")
}

var founderNames = map[string]bool{"datavorous": true, "nisarga": true}

func buildCards(entries []WebringEntry) string {
	var b strings.Builder
	for _, e := range entries {
		name := html.EscapeString(e.Name)
		b.WriteString(`<div class="card" data-name="`)
		b.WriteString(name)
		b.WriteString(`"><a class="card-link" href="`)
		b.WriteString(html.EscapeString(e.Url))
		b.WriteString(`" rel="noopener"><span class="avatar">`)
		b.WriteString(html.EscapeString(initial(e.Name)))
		if e.Gh != "" {
			b.WriteString(`<img src="https://avatars.githubusercontent.com/`)
			b.WriteString(html.EscapeString(e.Gh))
			b.WriteString(`?size=96" alt="" loading="lazy" onerror="this.setAttribute('data-failed','')" />`)
		}
		b.WriteString(`</span><span class="meta"><span class="name">`)
		b.WriteString(name)
		b.WriteString(`</span><span class="host">`)
		b.WriteString(html.EscapeString(host(e.Url)))
		b.WriteString(`</span></span><span class="arrow">&rarr;</span></a>`)
		if e.Gh != "" {
			gh := html.EscapeString(e.Gh)
			b.WriteString(`<a class="gh-link" href="https://github.com/`)
			b.WriteString(gh)
			b.WriteString(`" target="_blank" rel="noopener" aria-label="`)
			b.WriteString(name)
			b.WriteString(` on GitHub"><svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true" width="16" height="16"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg></a>`)
		}
		b.WriteString(`</div>`)
	}
	return b.String()
}

func renderIndex(founders, members []WebringEntry) []byte {
	out := strings.Replace(string(indexHTML), "<!--FOUNDERS-->", buildCards(founders), 1)
	out = strings.Replace(out, "<!--MEMBERS-->", buildCards(members), 1)
	out = strings.Replace(out, "<!--COUNT-->", fmt.Sprintf("%d members", len(founders)+len(members)), 1)
	return []byte(out)
}

func parsePost(data []byte) (BlogPost, error) {
	parts := bytes.SplitN(data, []byte("---\n"), 3)
	if len(parts) < 3 {
		return BlogPost{}, fmt.Errorf("invalid frontmatter")
	}
	var p BlogPost
	for _, line := range bytes.Split(parts[1], []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		kv := bytes.SplitN(line, []byte(": "), 2)
		if len(kv) != 2 {
			continue
		}
		key := string(bytes.TrimSpace(kv[0]))
		val := string(bytes.TrimSpace(kv[1]))
		switch key {
		case "title":
			p.Title = val
		case "date":
			p.Date = val
		case "slug":
			p.Slug = val
		case "external":
			p.External = val
		}
	}
	if p.Title == "" {
		return BlogPost{}, fmt.Errorf("missing title in frontmatter")
	}
	if p.Slug == "" {
		if p.External != "" {
			p.Slug = strings.ToLower(strings.ReplaceAll(p.Title, " ", "-"))
		} else {
			return BlogPost{}, fmt.Errorf("missing slug in frontmatter")
		}
	}
	p.Content = string(parts[2])
	return p, nil
}

func renderBlogIndex(posts []BlogPost) []byte {
	var b strings.Builder
	for _, p := range posts {
		href := "/blog/" + html.EscapeString(p.Slug)
		rel := ""
		if p.External != "" {
			href = html.EscapeString(p.External)
			rel = ` rel="noopener"`
		}
		b.WriteString(`<article class="post-card"><a href="`)
		b.WriteString(href)
		b.WriteString(`"`)
		b.WriteString(rel)
		b.WriteString(`><time datetime="`)
		b.WriteString(html.EscapeString(p.Date))
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(p.Date))
		b.WriteString(`</time><h2>`)
		b.WriteString(html.EscapeString(p.Title))
		if p.External != "" {
			b.WriteString(` <svg class="external-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" width="14" height="14"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>`)
		}
		b.WriteString(`</h2></a></article>`)
	}
	if len(posts) == 0 {
		b.WriteString(`<p class="empty-state">no posts yet.</p>`)
	}
	out := strings.Replace(string(blogHTML), "<!--POSTS-->", b.String(), 1)
	return []byte(out)
}

func renderPostPage(p BlogPost) []byte {
	out := strings.Replace(string(postHTML), "<!--TITLE-->", html.EscapeString(p.Title), 1)
	out = strings.Replace(out, "<!--DATE-->", html.EscapeString(p.Date), 1)
	out = strings.Replace(out, "<!--CONTENT-->", p.Content, 1)
	return []byte(out)
}

func main() {
	var webring []WebringEntry
	if err := json.Unmarshal(webringRaw, &webring); err != nil {
		slog.Error("failed to unmarshal webring json file", "error", err)
		os.Exit(1)
	}

	var founders, members []WebringEntry
	for _, e := range webring {
		if founderNames[e.Name] {
			founders = append(founders, e)
		} else {
			members = append(members, e)
		}
	}

	page := renderIndex(founders, members)

	md := goldmark.New()

	entries, err := postsFS.ReadDir("posts")
	if err != nil {
		slog.Error("failed to read posts directory", "error", err)
		os.Exit(1)
	}

	var blogPosts []BlogPost
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := postsFS.ReadFile("posts/" + entry.Name())
		if err != nil {
			slog.Warn("failed to read post file", "file", entry.Name(), "error", err)
			continue
		}
		p, err := parsePost(data)
		if err != nil {
			slog.Warn("skipping post", "file", entry.Name(), "error", err)
			continue
		}
		var buf bytes.Buffer
		if err := md.Convert([]byte(p.Content), &buf); err != nil {
			slog.Warn("failed to render markdown", "file", entry.Name(), "error", err)
			continue
		}
		p.Content = buf.String()
		blogPosts = append(blogPosts, p)
	}

	sort.Slice(blogPosts, func(i, j int) bool {
		return blogPosts[i].Date > blogPosts[j].Date
	})

	blogIndex := renderBlogIndex(blogPosts)

	blogPostMap := make(map[string]BlogPost, len(blogPosts))
	blogPostPages := make(map[string][]byte, len(blogPosts))
	for _, p := range blogPosts {
		blogPostMap[p.Slug] = p
		if p.External == "" {
			blogPostPages[p.Slug] = renderPostPage(p)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqHost := r.Host

		if strings.HasSuffix(reqHost, ".seggs.lol") && !slices.Contains(hostsToIgnore, reqHost) {
			sub := strings.TrimSuffix(reqHost, ".seggs.lol")
			for _, entry := range webring {
				if entry.Name == sub {
					http.Redirect(w, r, entry.Url, http.StatusFound)
					return
				}
			}

			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})

	http.HandleFunc("/webring", func(w http.ResponseWriter, r *http.Request) {
		buildJsonResponse(w, http.StatusOK, webring)
	})

	http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		if from == "" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "missing `from` query parameter",
			})
			return
		}

		dir := r.URL.Query().Get("dir")
		if dir == "" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "missing `dir` query parameter",
			})
			return
		}

		if dir != "next" && dir != "prev" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid `dir` query parameter. it can be either `next` or `prev` only",
			})
			return
		}

		index := -1
		for i, v := range webring {
			if v.Name == from {
				index = i
				break
			}
		}

		if index == -1 {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid `from` query parameter. can't find any webring entry's name as `from`",
			})
			return
		}

		url := ""

		if dir == "prev" {
			if index == 0 {
				url = webring[len(webring)-1].Url
			} else {
				url = webring[index-1].Url
			}
		} else {
			if index == len(webring)-1 {
				url = webring[0].Url
			} else {
				url = webring[index+1].Url
			}
		}

		setCorsHeaders(w)
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		setCorsHeaders(w)
		index := rand.Intn(len(webring))
		http.Redirect(w, r, webring[index].Url, http.StatusFound)
	})

	http.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/blog/")
		slug = strings.Trim(slug, "/")
		if slug == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(blogIndex)
			return
		}
		post, ok := blogPostMap[slug]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if post.External != "" {
			http.Redirect(w, r, post.External, http.StatusFound)
			return
		}
		postPage, ok := blogPostPages[slug]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(postPage)
	})

	workers.Serve(nil)
	fmt.Println("server is up and running at :8080")
}

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func buildJsonResponse(w http.ResponseWriter, statusCode int, v any) {
	setCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}
