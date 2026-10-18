kubectl -n ar-resourcenames create role flags-editor --verb=get,update --resource=configmaps --resource-name=feature-flags
kubectl -n ar-resourcenames create rolebinding toggler-flags --role=flags-editor --serviceaccount=ar-resourcenames:toggler
