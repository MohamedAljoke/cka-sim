kubectl -n ar-helm-owner get deployment checkout -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}' > /opt/course/ar-helm-owner/release.txt
echo 'helm ls -A' > /opt/course/ar-helm-owner/releases.sh
