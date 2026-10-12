conf=/etc/cni/net.d/10-kindnet.conflist
grep -q '"type": "ptpp"' "$conf" || exit 0
sed -i 's/"type": "ptpp"/"type": "ptp"/' "$conf"
