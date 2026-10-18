kubectl create clusterrole node-viewer --verb=get,list --resource=nodes,persistentvolumes
kubectl create clusterrolebinding node-viewer --clusterrole=node-viewer --group=auditors
kubectl auth can-i list nodes --as=someone --as-group=auditors
