`kubectl top pod` adds up all containers of a Pod. `--containers` splits the number per container:

```sh
kubectl -n tr-top-containers top pod app --containers
```

The columns are `POD`, `NAME` (the container), `CPU(cores)` and `MEMORY(bytes)`. `--sort-by` sorts
Pods, not the containers inside one, so pick the biggest by reading the list or sorting it with
`awk` and `sort -n`:

```sh
kubectl -n tr-top-containers top pod app --containers --no-headers \
  | awk '{print $4 + 0, $2}' | sort -n | tail -1
```

`awk`'s `+ 0` turns `52Mi` into `52`, which `sort -n` can compare. Plain `sort -h` doesn't know
the `Mi` suffix.
