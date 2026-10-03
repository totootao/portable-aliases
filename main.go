package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Entry 表示一条别名或函数定义
type Entry struct {
	Type  string `json:"type"`  // alias | function
	Name  string `json:"name"`
	Value string `json:"value"` // alias: 右侧命令; function: 函数体(可多行)
}

// Config 是 aliases.json 的结构, 作为权威数据源
type Config struct {
	Settings []string `json:"settings"`
	Entries  []Entry  `json:"entries"`
}

var (
	dir      string
	jsonPath string
	shPath   string
)

const shHeader = `#!/usr/bin/env bash
# =============================================================================
# aliases.sh  ——  由 aliases.json 自动生成, 请勿手改(保存时会被覆盖)
# 编辑请走 Web 管理界面, 或改 aliases.json 后运行: go run main.go generate
# 本文件可直接 source, 与 Docker 镜像共用, 保证可移植。
# =============================================================================
`

func load() (Config, error) {
	var c Config
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return c, err
	}
	return c, json.Unmarshal(data, &c)
}

// save 回写 json 并重新生成 aliases.sh, 保持可移植文件与配置一致
func save(c Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, b, 0644); err != nil {
		return err
	}
	return generate(c)
}

// generate 由 Config 生成可直接 source 的 aliases.sh
func generate(c Config) error {
	var b strings.Builder
	b.WriteString(shHeader)
	for _, s := range c.Settings {
		b.WriteString(s + "\n")
	}
	b.WriteString("\n")
	b.WriteString("# ------------------------- 别名与函数(受 Web 管理) -------------------------\n")
	for _, e := range c.Entries {
		if e.Type == "alias" {
			b.WriteString(fmt.Sprintf("alias %s='%s'\n", e.Name, e.Value))
		} else if e.Type == "function" {
			b.WriteString(e.Name + "() {\n")
			b.WriteString(e.Value + "\n")
			b.WriteString("}\n")
		}
	}
	return os.WriteFile(shPath, []byte(b.String()), 0644)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(dir, "index.html"))
}

func handleAliases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch r.Method {
	case http.MethodGet:
		c, err := load()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(c)

	case http.MethodPost:
		var e Entry
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "JSON 解析失败", 400)
			return
		}
		if e.Name == "" || e.Type == "" {
			http.Error(w, "name 与 type 必填", 400)
			return
		}
		c, err := load()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		found := false
		for i, x := range c.Entries {
			if x.Name == e.Name {
				c.Entries[i] = e
				found = true
				break
			}
		}
		if !found {
			c.Entries = append(c.Entries, e)
		}
		if err := save(c); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	default:
		http.Error(w, "仅支持 GET/POST", 405)
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodDelete {
		http.Error(w, "仅支持 DELETE", 405)
		return
	}
	name, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/api/aliases/"))
	if name == "" {
		http.Error(w, "name 必填", 400)
		return
	}
	c, err := load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	before := len(c.Entries)
	c.Entries = filterEntries(c.Entries, name)
	if len(c.Entries) == before {
		http.Error(w, "未找到 "+name, 404)
		return
	}
	if err := save(c); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func filterEntries(es []Entry, name string) []Entry {
	out := make([]Entry, 0, len(es))
	for _, e := range es {
		if e.Name != name {
			out = append(out, e)
		}
	}
	return out
}

func handleSh(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(shPath)
	if err != nil {
		fmt.Fprint(w, "# aliases.sh 尚未生成, 请先保存一次")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

func main() {
	dir = os.Getenv("APP_DIR")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	jsonPath = filepath.Join(dir, "aliases.json")
	shPath = filepath.Join(dir, "aliases.sh")

	// 子命令: go run main.go generate  —— 仅重新生成 aliases.sh 后退出
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		c, err := load()
		if err != nil {
			log.Fatal(err)
		}
		if err := generate(c); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("已生成 %s（共 %d 条）\n", shPath, len(c.Entries))
		return
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/aliases", handleAliases)
	http.HandleFunc("/api/aliases/", handleDelete)
	http.HandleFunc("/api/aliases.sh", handleSh)

	// 首次启动确保产物存在
	if _, err := os.Stat(shPath); os.IsNotExist(err) {
		if c, err2 := load(); err2 == nil {
			_ = generate(c)
		}
	}

	log.Println("别名 Web 管理面板已启动: http://0.0.0.0:8080  (无鉴权, 仅限本机/可信内网)")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
