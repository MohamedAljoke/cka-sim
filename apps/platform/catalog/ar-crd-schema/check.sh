crd() { kubectl get crd databases.cka.example.com -o json 2>/dev/null; }
v1() { crd | jq -c ".spec.versions[] | select(.name==\"v1\") | $1"; }
basics() { eq "Namespaced db databases" "$(crd | jq -r '[.spec.scope, .spec.names.shortNames[0], .spec.names.plural] | join(" ")')"; }
schema_ok() { eq '[["engine"],["mysql","postgres"],"integer",1]' "$(v1 '.schema.openAPIV3Schema.properties.spec | [.required, (.properties.engine.enum | sort), .properties.storageGB.type, .properties.storageGB.minimum]')"; }
columns() { eq '[["Engine",".spec.engine"],["Storage",".spec.storageGB"]]' "$(v1 '[.additionalPrinterColumns[] | [.name, .jsonPath]] | sort')"; }
orders() { eq "postgres 10" "$(kubectl -n ar-crd-schema get db orders -o jsonpath='{.spec.engine} {.spec.storageGB}' 2>/dev/null)"; }
rejected() { grep -qi 'unsupported value' "$COURSE/rejected.txt" 2>/dev/null && ! kubectl -n ar-crd-schema get db legacy >/dev/null 2>&1; }

check 1 "The CRD is namespaced, plural databases, short name db" basics
check 2 "engine is a required enum of postgres|mysql; storageGB an integer >= 1" schema_ok
check 1 "Printer columns Engine and Storage are defined" columns
check 1 "Database orders exists with postgres and 10" orders
check 1 "rejected.txt holds the enum error and legacy doesn't exist" rejected
