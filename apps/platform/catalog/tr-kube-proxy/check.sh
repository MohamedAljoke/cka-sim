NS=tr-kube-proxy
proxies() { kubectl -n kube-system get daemonset kube-proxy -o jsonpath='{.status.numberReady}/{.status.desiredNumberScheduled}'; }
everywhere() { eq "3/3" "$(proxies)"; }
serves() { [[ "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://web.$NS 2>/dev/null)" == *"Welcome to nginx"* ]]; }
untouched() { serves && eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.metadata.generation}')" && eq "80" "$(kubectl -n $NS get service web -o jsonpath='{.spec.ports[0].port}')"; }
if [ -n "$(kubectl -n kube-system get pods -l k8s-app=kube-proxy -o name)" ]; then
  wait_for 120 everywhere
  wait_for 30 serves
fi

check 3 "kube-proxy runs and is Ready on all 3 nodes" everywhere
check 3 "http://web.tr-kube-proxy serves the nginx page from Pod client" serves
check 1 "Service web and Deployment web were not modified" untouched
