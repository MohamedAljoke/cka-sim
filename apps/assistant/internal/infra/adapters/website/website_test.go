package website

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestAPageIsCutAtItsHeadingsWithTheirTrailAndAnchor(t *testing.T) {
	pages, err := Pages(clone(map[string]string{"docs/concepts/pdb.md": pdb}))
	if err != nil {
		t.Fatal(err)
	}

	chunks := pages[0].Chunks
	if len(chunks) != 3 {
		t.Fatalf("%d chunks: %+v", len(chunks), chunks)
	}
	if chunks[1].URL != "https://kubernetes.io/docs/concepts/pdb/#minavailable" ||
		!strings.HasPrefix(chunks[1].Text, "Disruptions › Fields › minAvailable\n\n") {
		t.Errorf("chunk %+v", chunks[1])
	}
	if chunks[2].URL != "https://kubernetes.io/docs/concepts/pdb/#how-to-drain" {
		t.Errorf("custom anchor: %s", chunks[2].URL)
	}
}

func TestHugoShortcodesBecomeText(t *testing.T) {
	pages, err := Pages(clone(map[string]string{"docs/concepts/pdb.md": pdb}))
	if err != nil {
		t.Fatal(err)
	}

	text := pages[0].Chunks[0].Text
	for _, want := range []string{"A Pod disruption budget", "Note: ", "kind: PodDisruptionBudget", "v1.37"} {
		if !strings.Contains(text, want) {
			t.Errorf("no %q in %q", want, text)
		}
	}
	if strings.Contains(text, "{{") || strings.Contains(text, "](/docs") {
		t.Errorf("shortcode or link target left in %q", text)
	}
}

func TestAHeadingInACodeBlockDoesNotCutThePage(t *testing.T) {
	pages, err := Pages(clone(map[string]string{"docs/tasks/script.md": script}))
	if err != nil {
		t.Fatal(err)
	}

	if len(pages[0].Chunks) != 1 || !strings.Contains(pages[0].Chunks[0].Text, "# not a heading") {
		t.Errorf("chunks %+v", pages[0].Chunks)
	}
}

func TestTheHashChangesOnlyWithThePage(t *testing.T) {
	first, _ := Pages(clone(map[string]string{"docs/concepts/pdb.md": pdb, "docs/tasks/script.md": script}))
	edited, _ := Pages(clone(map[string]string{
		"docs/concepts/pdb.md": strings.Replace(pdb, "voluntary", "planned", 1), "docs/tasks/script.md": script,
	}))

	if first[0].Hash == edited[0].Hash {
		t.Error("an edited page kept its hash")
	}
	if first[1].Hash != edited[1].Hash {
		t.Error("an unchanged page got a new hash")
	}
}

func TestAnExampleFileChangesThePageThatShowsIt(t *testing.T) {
	first, _ := Pages(clone(map[string]string{"docs/concepts/pdb.md": pdb}))
	files := clone(map[string]string{"docs/concepts/pdb.md": pdb})
	files["examples/policy/pdb.yaml"] = &fstest.MapFile{Data: []byte("kind: PodDisruptionBudget\nspec: {}\n")}
	changed, _ := Pages(files)

	if first[0].Hash == changed[0].Hash {
		t.Error("the page kept its hash after its example changed")
	}
}

func TestDraftsSkippedSectionsAndIndexPagesAreHandled(t *testing.T) {
	pages, err := Pages(clone(map[string]string{
		"docs/concepts/draft.md":                        "---\ntitle: Draft\ndraft: true\n---\n" + filler,
		"docs/contribute/style.md":                      "---\ntitle: Style\n---\n" + filler,
		"docs/reference/kubernetes-api/pod.md":          "---\ntitle: Pod\n---\n" + filler,
		"docs/concepts/workloads/_index.md":             "---\ntitle: Workloads\n---\n" + filler,
		"docs/concepts/workloads/pods/notes.txt":        filler,
		"docs/concepts/workloads/pods/no-title.md":      filler,
		"docs/concepts/workloads/pods/pod-lifecycle.md": "---\ntitle: Pod Lifecycle\n---\n" + filler,
	}))
	if err != nil {
		t.Fatal(err)
	}

	var urls []string
	for _, page := range pages {
		urls = append(urls, page.Chunks[0].URL)
	}
	if strings.Join(urls, " ") != "https://kubernetes.io/docs/concepts/workloads/ "+
		"https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/" {
		t.Errorf("urls %q", urls)
	}
}

func TestALongSectionIsCutIntoPieces(t *testing.T) {
	body := strings.Repeat(filler+"\n\n", 40)

	pages, err := Pages(clone(map[string]string{"docs/concepts/long.md": "---\ntitle: Long\n---\n" + body}))
	if err != nil {
		t.Fatal(err)
	}

	if len(pages[0].Chunks) < 2 {
		t.Fatalf("%d chunks", len(pages[0].Chunks))
	}
	for _, chunk := range pages[0].Chunks {
		if len(chunk.Text) > pieceMost+100 {
			t.Errorf("a chunk of %d characters", len(chunk.Text))
		}
	}
}

const filler = "A paragraph long enough to be kept as a chunk of its own, since shorter ones are dropped as noise."

const pdb = `---
title: Disruptions
---
A {{< glossary_tooltip text="Pod disruption budget" term_id="pdb" >}} limits voluntary disruptions in
[the cluster](/docs/concepts/cluster/). It works on {{< param "version" >}}.

{{< note >}}
Only voluntary disruptions count.
{{< /note >}}

{{% code_sample file="policy/pdb.yaml" %}}

## Fields

### minAvailable

The number of pods that must stay available during a drain, as a count or a percentage of the pods.

## How to drain {#how-to-drain}

Run kubectl drain on the node; the eviction API refuses evictions that would break the budget.
`

const script = "---\ntitle: Script\n---\n" + filler + "\n\n```bash\n# not a heading\necho hi\n```\n"

func clone(pages map[string]string) fstest.MapFS {
	files := fstest.MapFS{
		"examples/policy/pdb.yaml": {Data: []byte("kind: PodDisruptionBudget\n")},
	}
	for name, text := range pages {
		files[name] = &fstest.MapFile{Data: []byte(text)}
	}

	return files
}
