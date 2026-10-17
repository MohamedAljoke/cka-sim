can() { eq yes "$(kubectl auth can-i "$@" -n ar-subresources --as=someone --as-group=support 2>/dev/null)"; }
pods() { can get pods && can list pods; }
logs_exec() { can get pods --subresource=log && can create pods --subresource=exec; }
scale() { can update deployments --subresource=scale && can patch deployments --subresource=scale; }
nothing_more() { pods && scale && ! can update deployments && ! can patch deployments && ! can delete pods && ! can create pods && ! can get secrets; }

check 1 "support can get and list Pods" pods
check 2 "support can read logs and exec into Pods" logs_exec
check 2 "support can update and patch deployments/scale" scale
check 1 "support can't edit Deployments, delete or create Pods, or read Secrets" nothing_more
