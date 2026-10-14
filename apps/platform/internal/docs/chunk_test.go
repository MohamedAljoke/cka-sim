package docs

import (
	"strings"
	"testing"
	"testing/fstest"
)

const etcdPage = `---
title: Operating etcd clusters for Kubernetes
content_type: task
---

<!-- overview -->

{{< glossary_definition term_id="etcd" length="all" >}}

etcd is a consistent and highly-available key value store used as the {{< glossary_tooltip text="backing store" term_id="etcd" >}}
for all cluster data, on Kubernetes v{{< skew currentVersion >}}.

## {{% heading "prerequisites" %}}

You need a cluster with etcd. See the [etcd documentation](https://etcd.io/docs/) for more context on running it.

## Backing up an etcd cluster {#backing-up}

{{< note >}}
Take snapshots often, and keep them somewhere other than the control plane node itself. ![diagram](/images/etcd.svg)
{{< /note >}}

### Snapshot using etcdctl options

{{% code_sample file="admin/etcd-backup.yaml" %}}

` + "```shell\n# not a heading\nETCDCTL_API=3 etcdctl snapshot save snapshot.db\n```\n" + `
{{< mermaid >}}
graph TD; A-->B
{{< /mermaid >}}
`

func TestChunksFollowThePagesHeadingsWithLinksToEachSection(t *testing.T) {
	chunks, err := Chunks(fstest.MapFS{
		"docs/tasks/administer-cluster/configure-upgrade-etcd.md": {Data: []byte(etcdPage)},
		"examples/admin/etcd-backup.yaml":                         {Data: []byte("apiVersion: v1\nkind: Pod\n")},
	})
	if err != nil {
		t.Fatal(err)
	}

	page := "https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/"
	want := []struct{ url, start string }{
		{page, "Operating etcd clusters for Kubernetes\n\netcd is a consistent"},
		{page + "#before-you-begin", "Operating etcd clusters for Kubernetes › Before you begin\n\nYou need a cluster with etcd."},
		{page + "#backing-up", "Operating etcd clusters for Kubernetes › Backing up an etcd cluster\n\nNote: \nTake snapshots often,"},
		{page + "#snapshot-using-etcdctl-options",
			"Operating etcd clusters for Kubernetes › Backing up an etcd cluster › Snapshot using etcdctl options\n\n```yaml"},
	}
	if len(chunks) != len(want) {
		t.Fatalf("got %d chunks: %+v", len(chunks), chunks)
	}
	for i, w := range want {
		if chunks[i].URL != w.url || !strings.HasPrefix(chunks[i].Text, w.start) {
			t.Errorf("chunk %d = %s %q, want %s %q", i, chunks[i].URL, chunks[i].Text, w.url, w.start)
		}
		if chunks[i].Title != "Operating etcd clusters for Kubernetes" || chunks[i].Hash == "" {
			t.Errorf("chunk %d = %+v", i, chunks[i])
		}
	}
	all := ""
	for _, c := range chunks {
		all += c.Text + "\n"
	}
	for _, s := range []string{"backing store\nfor all cluster data, on Kubernetes v1.37.", "See the etcd documentation for more",
		"# admin/etcd-backup.yaml\napiVersion: v1\nkind: Pod", "# not a heading\nETCDCTL_API=3"} {
		if !strings.Contains(all, s) {
			t.Errorf("missing %q in\n%s", s, all)
		}
	}
	for _, s := range []string{"{{", "<!--", "](", "graph TD", "diagram", "glossary"} {
		if strings.Contains(all, s) {
			t.Errorf("left %q in\n%s", s, all)
		}
	}
}

func TestChunksGiveEachChunkItsOwnIDAndSkipWhatTheExamDoesNotNeed(t *testing.T) {
	section := "## Part\n\n" + strings.Repeat("Pods run containers. ", 10) + "\n"
	chunks, err := Chunks(fstest.MapFS{
		"docs/concepts/workloads/pods/_index.md":           {Data: []byte("---\ntitle: Pods\n---\n" + section + section)},
		"docs/concepts/empty/_index.md":                    {Data: []byte("---\ntitle: Empty\n---\n")},
		"docs/concepts/draft.md":                           {Data: []byte("---\ntitle: Draft\ndraft: true\n---\n" + section)},
		"docs/contribute/style.md":                         {Data: []byte("---\ntitle: Style\n---\n" + section)},
		"docs/reference/kubernetes-api/workload/pod-v1.md": {Data: []byte("---\ntitle: Pod\n---\n" + section)},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 2 || chunks[0].ID == chunks[1].ID {
		t.Fatalf("chunks = %+v", chunks)
	}
	if chunks[0].URL != "https://kubernetes.io/docs/concepts/workloads/pods/#part" ||
		chunks[0].ID != "docs/concepts/workloads/pods/_index.md#0" {
		t.Errorf("first chunk = %+v", chunks[0])
	}
}

func TestPiecesSplitALongSectionAtBlankLinesButNeverInsideCode(t *testing.T) {
	paragraph := strings.Repeat("word ", 300)
	code := "```\n" + strings.Repeat("line\n\n", 300) + "```"
	got := pieces(strings.Split(paragraph+"\n\n"+paragraph+"\n\n"+code, "\n"))

	if len(got) != 3 {
		t.Fatalf("got %d pieces", len(got))
	}
	if !strings.HasPrefix(got[2], "```") || !strings.HasSuffix(got[2], "```") {
		t.Errorf("the code block was split: %q", got[2][:20])
	}
}

func TestPiecesCutAHugeTableIntoPiecesTheModelCanRead(t *testing.T) {
	row := "| metric_name | " + strings.Repeat("x", 90) + " |"
	var table []string
	for range 200 {
		table = append(table, row)
	}
	got := pieces(table)

	if len(got) < 4 {
		t.Fatalf("got %d pieces", len(got))
	}
	for _, p := range got {
		if len(p) > pieceMost || !strings.HasPrefix(p, "| metric_name") {
			t.Errorf("piece of %d chars starts %q", len(p), p[:20])
		}
	}
}

func TestSkewFollowsTheDocsVersion(t *testing.T) {
	for args, want := range map[string]string{
		"currentVersion":                "1.37",
		"prevMinorVersion":              "1.36",
		"nextMinorVersion":              "1.38",
		`currentVersionAddMinor -1 "-"`: "1-36",
		"currentVersionAddMinor 1":      "1.38",
		"somethingNew":                  "",
	} {
		if got := skew(arguments(args).positional); got != want {
			t.Errorf("skew %s = %q, want %q", args, got, want)
		}
	}
}
