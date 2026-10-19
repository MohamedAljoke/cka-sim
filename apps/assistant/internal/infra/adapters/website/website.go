// Package website reads the Kubernetes docs from a clone of kubernetes/website and cuts each page into chunks.
package website

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

const (
	site       = "https://kubernetes.io/"
	minor      = 37
	pieceChars = 2000
	pieceMost  = 6000
	leastChars = 80
	// chunking is part of every page's hash: a change to how pages are cut embeds them all again.
	chunking = "website/1"
)

// The generated API references are left out (kubectl explain covers fields better), and so is contributing.
var skipped = []string{"docs/contribute/", "docs/reference/kubernetes-api/", "docs/reference/config-api/"}

var (
	shortcode   = regexp.MustCompile(`\{\{[<%]\s*(/?)\s*([\w-]+)((?:[^}]|\}[^}])*?)\s*[>%]\}\}`)
	argument    = regexp.MustCompile(`(?:([\w-]+)=)?(?:"([^"]*)"|([^\s"]+))`)
	comment     = regexp.MustCompile(`(?s)<!--.*?-->`)
	image       = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	link        = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	blankLines  = regexp.MustCompile(`\n{3,}`)
	heading     = regexp.MustCompile(`^(#{1,3})\s+(.*?)\s*(?:\{#([\w-]+)\})?\s*$`)
	notInAnchor = regexp.MustCompile(`[^\p{L}\p{N}\s_-]`)
	dropped     = []*regexp.Regexp{pair("mermaid"), pair("comment")}
	headings    = map[string]string{
		"prerequisites": "Before you begin",
		"objectives":    "Objectives",
		"cleanup":       "Clean up",
		"whatsnext":     "What's next",
		"synopsis":      "Synopsis",
		"options":       "Options",
		"parentoptions": "Options inherited from parent commands",
		"examples":      "Examples",
		"seealso":       "See also",
		"envvars":       "Environment variables",
	}
	callouts = map[string]string{"note": "Note: ", "caution": "Caution: ", "warning": "Warning: "}
)

func pair(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?s)\{\{[<%]\s*` + name + `\b.*?\{\{[<%]\s*/\s*` + name + `\s*[>%]\}\}`)
}

// Pages reads content/en of a kubernetes/website clone: the pages under docs/, with the YAML of examples/
// inlined where a page shows it.
func Pages(root fs.FS) ([]docs.Page, error) {
	var pages []docs.Page
	err := fs.WalkDir(root, "docs", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || path.Ext(name) != ".md" || isSkipped(name) {
			return nil
		}

		raw, err := fs.ReadFile(root, name)
		if err != nil {
			return err
		}

		page, ok, err := pageOf(root, name, raw)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		if ok {
			pages = append(pages, page)
		}

		return nil
	})

	return pages, err
}

func isSkipped(name string) bool {
	for _, prefix := range skipped {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

type frontMatter struct {
	Title string `json:"title"`
	Draft bool   `json:"draft"`
}

type section struct {
	trail  string
	anchor string
	body   []string
}

func pageOf(root fs.FS, name string, raw []byte) (docs.Page, bool, error) {
	meta, body, err := split(raw)
	if err != nil {
		return docs.Page{}, false, err
	}

	if meta.Draft || meta.Title == "" {
		return docs.Page{}, false, nil
	}

	url := pageURL(name)
	hash := sha256.New()
	hash.Write([]byte(chunking))
	page := docs.Page{Path: name, Chunks: []docs.Chunk{}}

	for _, s := range sections(meta.Title, clean(root, body)) {
		for _, piece := range pieces(s.body) {
			if len(piece) < leastChars {
				continue
			}

			chunk := docs.Chunk{Title: meta.Title, URL: url, Text: s.trail + "\n\n" + piece}
			if s.anchor != "" {
				chunk.URL += "#" + s.anchor
			}

			hash.Write([]byte("\x00" + chunk.URL + "\x00" + chunk.Text))
			page.Chunks = append(page.Chunks, chunk)
		}
	}
	page.Hash = hex.EncodeToString(hash.Sum(nil))

	return page, true, nil
}

func split(raw []byte) (frontMatter, string, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")

	var meta frontMatter

	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return meta, text, nil
	}

	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return meta, text, nil
	}

	if err := yaml.Unmarshal([]byte(head), &meta); err != nil {
		return meta, "", fmt.Errorf("front matter: %w", err)
	}

	return meta, body, nil
}

func pageURL(name string) string {
	p := strings.TrimSuffix(name, ".md")
	if base := path.Base(p); base == "_index" || base == "index" {
		p = path.Dir(p)
	}

	return site + p + "/"
}

// clean turns Hugo's markdown into plain markdown: shortcodes resolved, comments, images and link targets gone.
func clean(root fs.FS, body string) string {
	body = comment.ReplaceAllString(body, "")
	for _, re := range dropped {
		body = re.ReplaceAllString(body, "")
	}
	body = shortcode.ReplaceAllStringFunc(body, func(match string) string {
		m := shortcode.FindStringSubmatch(match)
		if m[1] == "/" {
			return ""
		}

		return expand(root, m[2], arguments(m[3]))
	})
	body = image.ReplaceAllString(body, "")
	body = link.ReplaceAllString(body, "$1")

	return blankLines.ReplaceAllString(body, "\n\n")
}

type args struct {
	named      map[string]string
	positional []string
}

func arguments(s string) args {
	given := args{named: map[string]string{}}

	for _, m := range argument.FindAllStringSubmatch(s, -1) {
		value := m[2] + m[3]
		if m[1] != "" {
			given.named[m[1]] = value
		} else {
			given.positional = append(given.positional, value)
		}
	}

	return given
}

func expand(root fs.FS, name string, given args) string {
	switch name {
	case "glossary_tooltip":
		return given.named["text"]
	case "code_sample", "codenew":
		return example(root, given.named["file"], given.named["language"])
	case "skew":
		return skew(given.positional)
	case "param":
		if len(given.positional) > 0 && given.positional[0] == "version" {
			return fmt.Sprintf("v1.%d", minor)
		}

		return ""
	case "heading":
		if len(given.positional) > 0 {
			return headings[given.positional[0]]
		}

		return ""
	case "feature-state":
		if given.named["state"] == "" || given.named["for_k8s_version"] == "" {
			return ""
		}

		return fmt.Sprintf("FEATURE STATE: Kubernetes %s [%s]", given.named["for_k8s_version"], given.named["state"])
	case "tab":
		if given.named["name"] != "" {
			return given.named["name"] + ":\n"
		}

		return ""
	}

	return callouts[name]
}

func example(root fs.FS, file, language string) string {
	if file == "" {
		return ""
	}

	raw, err := fs.ReadFile(root, path.Join("examples", file))
	if err != nil {
		return ""
	}

	if language == "" {
		language = strings.TrimPrefix(path.Ext(file), ".")
	}

	return fmt.Sprintf("```%s\n# %s\n%s\n```", language, file, bytes.TrimSpace(raw))
}

func skew(positional []string) string {
	if len(positional) == 0 {
		return ""
	}

	version := minor

	switch positional[0] {
	case "currentVersion", "latestVersion", "currentPatchVersion":
	case "nextMinorVersion":
		version++
	case "prevMinorVersion":
		version--
	case "currentVersionAddMinor":
		if len(positional) < 2 {
			return ""
		}

		added, err := strconv.Atoi(positional[1])
		if err != nil {
			return ""
		}

		separator := "."
		if len(positional) > 2 {
			separator = positional[2]
		}

		return fmt.Sprintf("1%s%d", separator, minor+added)
	default:
		return ""
	}

	return fmt.Sprintf("1.%d", version)
}

// sections splits a page at its #, ## and ### headings, never inside a code block; each keeps the trail of
// headings above it and its anchor.
func sections(title, body string) []section {
	current := section{trail: title}

	var (
		out     []section
		chapter string
	)
	fenced := false

	for line := range strings.SplitSeq(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
		}

		m := heading.FindStringSubmatch(line)
		if fenced || m == nil {
			current.body = append(current.body, line)
			continue
		}

		out = append(out, current)
		text := strings.TrimSpace(m[2])
		trail := title + " › " + text

		switch len(m[1]) {
		case 1, 2:
			chapter = text
		case 3:
			if chapter != "" {
				trail = title + " › " + chapter + " › " + text
			}
		}

		anchor := m[3]
		if anchor == "" {
			anchor = autoAnchor(text)
		}
		current = section{trail: trail, anchor: anchor}
	}

	return append(out, current)
}

func autoAnchor(text string) string {
	text = notInAnchor.ReplaceAllString(strings.ToLower(text), "")

	return strings.Join(strings.Fields(text), "-")
}

// pieces joins a section's paragraphs into pieces of about pieceChars, and cuts any piece longer than pieceMost.
func pieces(lines []string) []string {
	var (
		paragraphs []string
		paragraph  []string
	)
	fenced := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
		}

		if trimmed == "" && !fenced {
			if len(paragraph) > 0 {
				paragraphs = append(paragraphs, strings.Join(paragraph, "\n"))
				paragraph = nil
			}

			continue
		}

		paragraph = append(paragraph, line)
	}

	if len(paragraph) > 0 {
		paragraphs = append(paragraphs, strings.Join(paragraph, "\n"))
	}

	var (
		out   []string
		piece strings.Builder
	)
	for _, p := range paragraphs {
		if piece.Len() > 0 && piece.Len()+len(p)+2 > pieceChars {
			out = append(out, piece.String())
			piece.Reset()
		}

		if piece.Len() > 0 {
			piece.WriteString("\n\n")
		}

		piece.WriteString(p)
	}

	if piece.Len() > 0 {
		out = append(out, piece.String())
	}

	capped := make([]string, 0, len(out))
	for _, p := range out {
		capped = append(capped, cut(p)...)
	}

	return capped
}

func cut(piece string) []string {
	var out []string

	for len(piece) > pieceMost {
		end := strings.LastIndexByte(piece[:pieceMost], '\n')
		if end <= 0 {
			end = pieceMost
		}
		out = append(out, piece[:end])
		piece = strings.TrimLeft(piece[end:], "\n")
	}

	return append(out, piece)
}
