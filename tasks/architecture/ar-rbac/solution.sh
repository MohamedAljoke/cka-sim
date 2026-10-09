k -n ci create serviceaccount deployer
k -n ci create role deployer --verb=get,list,create,update --resource=deployments
k -n ci create rolebinding deployer --role=deployer --serviceaccount=ci:deployer
on_host "kubectl --kubeconfig /etc/kubernetes/admin.conf auth can-i delete deployments -n ci --as=system:serviceaccount:ci:deployer > $(course_dir)/can-delete.txt || true"
