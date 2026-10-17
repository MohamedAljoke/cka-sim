HOST=legacy.nw-gateway-rewrite.test
filter() { kubectl -n nw-gateway-rewrite get httproute legacy -o jsonpath='{.spec.rules[0].filters[0].type} {.spec.rules[0].filters[0].urlRewrite.path.type}'; }

check 3 "/legacy/items reaches app as /items" eq "app /items" "$(fetch $HOST /legacy/items)"
check 1 "/legacy reaches app as /" eq "app /" "$(fetch $HOST /legacy)"
check 1 "The rule uses a URLRewrite filter with ReplacePrefixMatch" eq "URLRewrite ReplacePrefixMatch" "$(filter)"
