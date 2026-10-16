`KUBECONFIG` can list several files, separated by `:`. kubectl then merges them in memory: the
first file that sets a value wins, and the lists of clusters, users and contexts are combined.
`config view --flatten` writes that merged view out as one file, with every referenced file
embedded as `*-data`:

```sh
KUBECONFIG=team-a.kubeconfig:team-b.kubeconfig kubectl config view --flatten > merged.kubeconfig
kubectl --kubeconfig merged.kubeconfig config get-contexts
kubectl --kubeconfig merged.kubeconfig config use-context team-b                       # = kubectx team-b
kubectl --kubeconfig merged.kubeconfig config set-context --current --namespace=ar-merge-b   # = kubens ar-merge-b
```

Without `--flatten` and `--raw`, `config view` redacts the certificate data, and the "merged" file
wouldn't work. `kubectx` and `kubens` (from the course) are just shortcuts for these two `config`
commands. They aren't installed in the exam, so know the long forms.

Entry names must be unique across files: two files that both name their user `admin` would merge
into one entry, and the first file's credentials would win silently.
