# cka-sim assistant: the next improvements

Each section says what is wrong today (with the file), what to change, **how to test it**, and what the test
teaches you. The tests are the point: each one proves a behaviour instead of trusting the prompt.

The lab and infrastructure improvements live in [IMPROVEMENTS-LABS.md](IMPROVEMENTS-LABS.md).

## What decision-memory already does

These sections come from decision-memory (Transfeera study project): a Go MCP server that records
engineering decisions and searches them with pgvector plus Postgres full-text search, the same shape as
`search_docs`. Each section points at the piece of decision-memory it borrows.

## 1. `search_docs` always returns five passages, even bad ones

`apps/platform/internal/docs/store.go` fuses the vector and keyword ranks with RRF and returns the top `k`
(`shown = 5`). RRF only orders, it never rejects. Ask about Helm hooks or an AWS console and the assistant
still gets five kubernetes.io passages, and the tool description tells it to cite them. `Result` has no
score, so the assistant can't tell a strong match from a weak one.

decision-memory drops every hit whose cosine similarity is under `DM_MIN_SIMILARITY` (0.65, calibrated on
`gemini-embedding-001`) and returns an empty list when nothing passes.

### Change

- Keep RRF for the order. Also select `1 - (embedding <=> $1)` as `similarity`, return it in `Result`, and
  drop rows under `CKA_SIM_DOCS_MIN_SIMILARITY`. Keyword-only hits with a low similarity go too: they are
  the "the page says `port` too" matches.
- `FindBy` turns zero results into `ErrEmpty`, "the Kubernetes docs aren't indexed yet". With a threshold,
  empty means "nothing relevant", so split them: check that the index has rows (the `model` meta key is
  already there) and return an empty list, not an error, when it does.
- `search_docs` description: "An empty result means the docs don't cover this; say so and don't cite a
  link."

### Test it

- `docs_test.go` with the fake store: the threshold reaches the store; an empty index still gives
  `ErrEmpty`; a filled index with no match gives an empty list and no error. Dropping the rows under it is
  SQL in `Postgres.Search`, so the eval's numbers cover that.
- In the Assistant panel, ask "how do I write a Helm post-install hook?". Before: a cited link that doesn't
  answer it. After: it says the exam docs don't cover it.

You learn that ranking and relevance are different questions: RRF answers "which is best", a threshold
answers "is any of them good enough".

## 2. The docs eval only tests hits, and needs Vertex every run

`apps/platform/cmd/docs/eval.yaml` has 32 questions and each has a `want` page. Nothing checks that the
search stays quiet when it should, so §1's threshold would be untested. `evaluate` also embeds every
question through Vertex on every run and needs a filled index, so it can't run in CI.

decision-memory's `make eval` (`internal/domain/eval`) does three things worth copying:

- **Questions with no answer.** A question with no expected decision passes when the search returns
  nothing.
- **Cached embeddings.** `fixtures/embeddings.json` keeps each text's vector by model, so the eval runs
  offline and only calls the model for new or changed text.
- **A near-threshold report.** It lists the questions whose right answer scores close to the threshold
  (just above or just below), with a hint on what to fix. These are the ones the next model or chunking
  change breaks.

### Change

- Allow `want: []` in `eval.yaml` and add 8 to 10 questions outside the exam docs: Helm internals, a cloud
  console, Docker Compose, Terraform, a non-Kubernetes question. Report them as their own line: "quiet when
  it should be: 7 / 8".
- Cache the question vectors in `apps/platform/cmd/docs/eval-embeddings.json`, keyed by model and query
  text.
- Print each `want` page's similarity next to the threshold, and the closest wrong page.
- Optional, for `cka-sim up` without GCP: an Ollama `nomic-embed-text` embedder (also 768 dims) behind
  `llm.Embedder`. `docs index` already refuses to mix models through the `model` meta key. The threshold
  must be calibrated per model: nomic scores lower than Gemini for the same pair.

### Test it

```sh
make docs-eval                            # first run fills the cache
CKA_SIM_VERTEX_PROJECT= make docs-eval    # second run: same numbers, no Vertex call
```

Then lower the threshold with `CKA_SIM_DOCS_MIN_SIMILARITY=0.3`: the "no answer" line drops. Raise it to
0.9: recall drops. The value you keep is where both lines are acceptable, and the near-threshold list shows
which questions decide it.

You learn to measure a retrieval system in both directions, and that a threshold is a trade-off, not a
constant.

## 3. The assistant can quote docs text that isn't in the docs

The assistant explains with kubernetes.io links, and `guard.go` already gathers the sources from
`search_docs` output (`docsSources`). Nothing checks that a quoted passage actually came from those sources.
A wrong flag name in a quote looks exactly as trustworthy as a right one.

decision-memory's ingest (`internal/domain/ingest/service.go`) has an LLM pull decisions out of a markdown
file, each with a quote of the reason. `normalizeCandidates` keeps the quote only if it appears word for word
in the source, after collapsing whitespace. Otherwise it's dropped.

### Change

- In the chat, keep the text of every chunk `search_docs` returned. When the reply has a quote (a `>` block
  or a quoted sentence next to a kubernetes.io link), look for it in those chunks with whitespace collapsed.
  If it isn't there, strip the quote and keep the link, or mark it "paraphrase".
- Use the same check when you draft a task's `explain.md` with an LLM (the `new-task` skill): every quoted
  line must exist in the indexed page it links to.

### Test it

- `guard_test.go`: a reply quoting a chunk verbatim keeps it; one with a changed flag
  (`--etcd-servers` → `--etcd-server`) loses it; one with extra line breaks keeps it.
- Count how often it fires on a few real chats. If it's often, the system prompt should ask for links
  instead of quotes.

You learn grounding: checking the model's output against the source with plain string matching, not
another model.

## 4. Docs search ignores the task the learner has open

`current_task` knows the task's domain, but `search_docs` searches the whole site. "port" brings back
Services, probes and NetworkPolicy alike.

decision-memory tags each decision (`technical`, `infrastructure`, `security`…) and `search_decisions`
takes an optional tag filter.

### Change

- At index time, map each chunk's URL prefix to a CKA domain (`concepts/storage/` → Storage,
  `concepts/services-networking/` → Services & Networking, `tasks/debug/` → Troubleshooting…) and store it
  as a `domain` column. Pages that fit none (`reference/kubectl/`) stay untagged and always match.
- `search_docs` takes an optional `domain`. The tool description says to pass the open task's domain first
  and search without it if that finds nothing.

### Test it

- Add `domain` to some `eval.yaml` questions and compare recall and MRR with and without the filter: `make
  docs-eval` prints a `hybrid ≥ 0.65 + domain` row under the unfiltered ones. A miss only that row has points at
  the prefix table in `internal/docs/chunk.go`.
- Make sure the filter is off by default in `eval`, so the existing numbers stay comparable.
- After the migration, `make docs-index` prints `retagged N` and embeds nothing; a second run prints `retagged 0`.

## 5. The assistant forgets what each learner got wrong

Each chat starts from zero. A learner who mixed up `--from-literal` and `--from-file` last week gets no
reminder on the next Secrets task.

decision-memory's `record_decision` has the agent write the decision, **show the draft and save only after
the user approves or edits it** (long drafts open in `$DM_EDITOR`). `search_decisions` brings it back when
it's relevant.

### Change

- Two tools on the cka-sim MCP: `record_note` (task id, the mistake, the fix, a docs link) and
  `search_notes` (by meaning and keyword, filtered to the learner's account and optionally the domain).
  Notes live in Postgres with an embedding, per account.
- The assistant asks after a failed `run_check` or an `explain`: "want me to save a note on what tripped you
  up?", shows the note, and saves it only after the learner says yes.
- Before a mock exam, or when a task opens in a domain with notes, the assistant searches them and says
  "last time you…".
- The tool doesn't touch the cluster, like the other tools. Notes are the learner's own text, so treat what
  `search_notes` returns as data, the way the cluster output already is.

### Test it

- Unit: one account can't read another's notes; a note without approval isn't stored (the tool takes the
  approved text, the assistant prompt does the asking).
- By hand: fail `wl-secret`, save the note, open `wl-secret-decode`, and check that the assistant brings
  it up without being asked.

You learn agent memory: what to store, who approves it, and when to recall it.

## Order

1. §2's "no answer" questions and the embedding cache first, then §1's threshold measured against them.
2. §3 grounding and §4 domain filter next.
3. §5 last, since it adds storage and a new flow.
