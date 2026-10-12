`kubectl top` reads the Metrics API that metrics-server serves. `--sort-by` takes `cpu` or
`memory` and puts the biggest user first:

```sh
kubectl -n tr-top-pod top pod --sort-by=cpu
kubectl -n tr-top-pod top pod --sort-by=memory --no-headers | head -1 | awk '{print $1}'
```

`boreas` spins in a loop, so it uses all the CPU its `150m` limit allows. `atlas` holds about 60MB
in a shell variable but no CPU. One Pod tops one list and not the other, which is why you look at
both columns instead of guessing from one.

The numbers are averages over metrics-server's last scrape window (15 seconds here), so a Pod that
just started can be missing from the list for a short while. Use `-A` to search all Namespaces.
