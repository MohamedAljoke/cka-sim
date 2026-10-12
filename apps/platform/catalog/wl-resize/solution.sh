kubectl -n wl-resize patch pod worker --subresource resize -p '{"spec":{"containers":[{"name":"worker","resources":{"requests":{"cpu":"300m"},"limits":{"cpu":"400m"}}}]}}'
kubectl -n wl-resize get pod worker -o jsonpath='{.status.containerStatuses[0].resources}'
