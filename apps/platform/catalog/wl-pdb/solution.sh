kubectl -n wl-pdb create poddisruptionbudget api-pdb --selector=app=api --min-available=2
kubectl -n wl-pdb get pdb api-pdb
