kubectl -n ci create serviceaccount deployer
kubectl -n ci create role deployer --verb=get,list,create,update --resource=deployments
kubectl -n ci create rolebinding deployer --role=deployer --serviceaccount=ci:deployer
kubectl auth can-i delete deployments -n ci --as=system:serviceaccount:ci:deployer > "$COURSE/can-delete.txt" || true
