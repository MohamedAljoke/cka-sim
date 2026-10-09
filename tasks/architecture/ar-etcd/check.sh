# After a restore the kubelet restarts etcd and the API server flaps for a while, so every
# check waits for its condition instead of trusting the first answer.
api_up() { k get --raw /readyz >/dev/null; }
snapshot_ok() { on_host "etcdutl snapshot status $(course_dir)/etcd-snapshot.db"; }
marker_back() { k -n etcd-check get configmap marker; }
# NotFound, not just any error: a down API server must not count as "restored".
after_backup_gone() { [[ "$(k -n etcd-check get configmap created-after-backup 2>&1)" == *NotFound* ]]; }

restored_and_up() { marker_back && api_up; }
wait_for 90 marker_back && wait_for 60 after_backup_gone
wait_for 60 api_up
check 3 "etcd-snapshot.db is a valid etcd snapshot" snapshot_ok
check 3 "ConfigMap etcd-check/marker is back" marker_back
check 1 "objects created after the old backup are gone" after_backup_gone
check 1 "API server is healthy after the restore" restored_and_up
