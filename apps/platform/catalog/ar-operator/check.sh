NS=ar-operator
operator() { eq 1 "$(kubectl -n web-operator get deployment web-operator -o jsonpath='{.status.availableReplicas}')"; }
webapp() { eq "nginx:1.27-alpine 2 80" "$(kubectl -n $NS get webapp shop -o jsonpath='{.spec.image} {.spec.replicas} {.spec.port}')"; }
built() { eq "WebApp shop 2" "$(kubectl -n $NS get deployment shop -o jsonpath='{.metadata.ownerReferences[0].kind} {.metadata.ownerReferences[0].name} {.status.readyReplicas}')"; }
serves() { built && curl -s -m 3 "http://$(kubectl -n $NS get service shop -o jsonpath='{.spec.clusterIP}')" | grep -q 'Welcome to nginx'; }
operator && webapp && wait_for 60 built

check 1 "The web-operator Deployment is available" operator
check 1 "WebApp shop has image nginx:1.27-alpine, 2 replicas and port 80" webapp
check 2 "Deployment shop is owned by WebApp shop and has 2 ready replicas" built
check 1 "Service shop, built by the operator, answers with nginx" serves
