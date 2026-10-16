package docs

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
)

const (
	Site       = "https://kubernetes.io/"
	Minor      = 37
	pieceChars = 2000
	pieceMost  = 6000
	leastChars = 80
)

var skipped = []string{"docs/contribute/", "docs/reference/kubernetes-api/", "docs/reference/config-api/"}

var domains = map[string]string{
	"concepts/storage/": "storage",
	"tasks/configure-pod-container/configure-persistent-volume-storage/": "storage",
	"tasks/configure-pod-container/configure-projected-volume-storage/":  "storage",
	"tasks/configure-pod-container/configure-volume-storage/":            "storage",
	"concepts/services-networking/":                                      "networking",
	"tasks/network/":                                                     "networking",
	"tasks/access-application-cluster/":                                  "networking",
	"tasks/administer-cluster/coredns/":                                  "networking",
	"tasks/administer-cluster/dns-":                                      "networking",
	"tasks/administer-cluster/declare-network-policy/":                   "networking",
	"concepts/workloads/":                                                "workloads",
	"concepts/scheduling-eviction/":                                      "workloads",
	"concepts/configuration/":                                            "workloads",
	"concepts/policy/":                                                   "workloads",
	"tasks/run-application/":                                             "workloads",
	"tasks/configure-pod-container/":                                     "workloads",
	"tasks/inject-data-application/":                                     "workloads",
	"tasks/configmap-secret/":                                            "workloads",
	"tasks/job/":                                                         "workloads",
	"tasks/administer-cluster/manage-resources/":                         "workloads",
	"setup/":                           "architecture",
	"concepts/architecture/":           "architecture",
	"concepts/security/":               "architecture",
	"concepts/extend-kubernetes/":      "architecture",
	"concepts/cluster-administration/": "architecture",
	"reference/access-authn-authz/":    "architecture",
	"reference/setup-tools/":           "architecture",
	"tasks/administer-cluster/":        "architecture",
	"tasks/extend-kubernetes/":         "architecture",
	"tasks/tls/":                       "architecture",
	"tasks/manage-kubernetes-objects/": "architecture",
	"tasks/configure-pod-container/configure-service-account/":             "architecture",
	"tasks/access-application-cluster/configure-access-multiple-clusters/": "architecture",
	"tasks/debug/": "troubleshooting",
}

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

type Chunk struct {
	ID     string
	URL    string
	Title  string
	Text   string
	Hash   string
	Domain string
}

type page struct {
	Title string `json:"title"`
	Draft bool   `json:"draft"`
}

type section struct {
	trail  string
	anchor string
	body   []string
}

func Chunks(root fs.FS) ([]Chunk, error) {
	var chunks []Chunk
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
		found, err := pageChunks(root, name, raw)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		chunks = append(chunks, found...)
		return nil
	})
	return chunks, err
}

func isSkipped(name string) bool {
	for _, prefix := range skipped {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func pageChunks(root fs.FS, name string, raw []byte) ([]Chunk, error) {
	meta, body, err := frontMatter(raw)
	if err != nil {
		return nil, err
	}
	if meta.Draft || meta.Title == "" {
		return nil, nil
	}
	url := pageURL(name)
	domain := domainOf(url)
	var chunks []Chunk
	for _, s := range sections(meta.Title, clean(root, body)) {
		for _, piece := range pieces(s.body) {
			if len(piece) < leastChars {
				continue
			}
			c := Chunk{
				ID:     fmt.Sprintf("%s#%d", name, len(chunks)),
				URL:    url,
				Title:  meta.Title,
				Text:   s.trail + "\n\n" + piece,
				Domain: domain,
			}
			if s.anchor != "" {
				c.URL += "#" + s.anchor
			}
			sum := sha256.Sum256([]byte(c.URL + "\n" + c.Text))
			c.Hash = hex.EncodeToString(sum[:])
			chunks = append(chunks, c)
		}
	}
	return chunks, nil
}

func domainOf(url string) string {
	page := strings.TrimPrefix(url, Site+"docs/")
	best, domain := "", ""
	for prefix, d := range domains {
		if strings.HasPrefix(page, prefix) && len(prefix) > len(best) {
			best, domain = prefix, d
		}
	}
	return domain
}

func frontMatter(raw []byte) (page, string, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	var meta page
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
	return Site + p + "/"
}

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
	a := args{named: map[string]string{}}
	for _, m := range argument.FindAllStringSubmatch(s, -1) {
		value := m[2] + m[3]
		if m[1] != "" {
			a.named[m[1]] = value
		} else {
			a.positional = append(a.positional, value)
		}
	}
	return a
}

func expand(root fs.FS, name string, a args) string {
	switch name {
	case "glossary_tooltip":
		return a.named["text"]
	case "code_sample", "codenew":
		return example(root, a.named["file"], a.named["language"])
	case "skew":
		return skew(a.positional)
	case "param":
		if len(a.positional) > 0 && a.positional[0] == "version" {
			return fmt.Sprintf("v1.%d", Minor)
		}
		return ""
	case "heading":
		if len(a.positional) > 0 {
			return headings[a.positional[0]]
		}
		return ""
	case "feature-state":
		if a.named["state"] == "" || a.named["for_k8s_version"] == "" {
			return ""
		}
		return fmt.Sprintf("FEATURE STATE: Kubernetes %s [%s]", a.named["for_k8s_version"], a.named["state"])
	case "tab":
		if a.named["name"] != "" {
			return a.named["name"] + ":\n"
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
	minor := Minor
	switch positional[0] {
	case "currentVersion", "latestVersion", "currentPatchVersion":
	case "nextMinorVersion":
		minor++
	case "prevMinorVersion":
		minor--
	case "currentVersionAddMinor":
		if len(positional) < 2 {
			return ""
		}
		n, err := strconv.Atoi(positional[1])
		if err != nil {
			return ""
		}
		separator := "."
		if len(positional) > 2 {
			separator = positional[2]
		}
		return fmt.Sprintf("1%s%d", separator, Minor+n)
	default:
		return ""
	}
	return fmt.Sprintf("1.%d", minor)
}

func sections(title, body string) []section {
	current := section{trail: title}
	var out []section
	var h2 string
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
			h2 = text
		case 3:
			if h2 != "" {
				trail = title + " › " + h2 + " › " + text
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

func pieces(lines []string) []string {
	var paragraphs []string
	var paragraph []string
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
	var out []string
	var piece strings.Builder
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
	var capped []string
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
