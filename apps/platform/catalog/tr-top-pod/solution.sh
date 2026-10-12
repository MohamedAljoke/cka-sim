kubectl -n tr-top-pod top pod --sort-by=cpu --no-headers | head -1 | awk '{print $1}' > /opt/course/tr-top-pod/cpu.txt
kubectl -n tr-top-pod top pod --sort-by=memory --no-headers | head -1 | awk '{print $1}' > /opt/course/tr-top-pod/memory.txt
