Each Pod has its own **network namespace**: its own interfaces, IPs, routes and iptables. The CNI
connects it to the node with a **veth pair**, two virtual interfaces joined like a cable: `eth0`
inside the Pod and `veth...` on the node. The node routes the Pod's IP into its end of the cable.

```sh
crictl ps --name probe                              # container ID
crictl inspect <id> | jq .info.pid                  # its host PID
nsenter -t <pid> -n ip addr                         # run `ip` inside the Pod's network namespace
#   2: eth0@if7: ...   inet 10.244.2.9/24            <- "@if7": the peer is ifindex 7 on the host
ip -o link | grep '^7:'                             # 7: vethd2a1b3c4@if2: ...
ip route | grep 10.244.2.9                          # 10.244.2.9 dev vethd2a1b3c4 scope host
```

`nsenter -n` enters just the network namespace, so you get the node's tools (`ip`, `ss`, `tcpdump`)
with the Pod's view of the network, which is handy for minimal images with no tools inside.
`ip netns` won't list these namespaces: runtimes don't register them under `/var/run/netns`.

Every container in the Pod, including the `pause` sandbox container that holds the namespace,
shares this one `eth0`. That's why containers in a Pod reach each other on `localhost`.
