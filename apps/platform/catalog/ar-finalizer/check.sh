NS=ar-finalizer
pod_gone() { ! kubectl -n $NS get pod legacy-app >/dev/null 2>&1; }
pvc_gone() { ! kubectl -n $NS get pvc data >/dev/null 2>&1; }
clean() { pod_gone && pvc_gone; }
pod_gone && wait_for 60 pvc_gone

check 2 "Pod legacy-app, which held the PVC, is deleted" pod_gone
check 3 "PVC data is gone" clean
