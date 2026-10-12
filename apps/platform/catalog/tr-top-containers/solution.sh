kubectl -n tr-top-containers top pod app --containers --no-headers | awk '{print $4 + 0, $2}' | sort -n | tail -1 | cut -d' ' -f2 > /opt/course/tr-top-containers/container.txt
