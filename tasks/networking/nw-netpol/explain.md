**Concept.** Pods accept all traffic until **some** NetworkPolicy selects them. The moment one
policy with `policyTypes: [Ingress]` selects a Pod, only what the policies' `ingress` rules
allow gets in — everything else is denied. Policies only add allowances; there is no "deny" rule.

**The trap is the YAML shape of `from`.** Each `-` item is OR-ed:

```yaml
from:
- podSelector: {matchLabels: {app: frontend}}          # OR
- namespaceSelector: {matchLabels: {kubernetes.io/metadata.name: monitoring}}
```

Putting both selectors in **one** item (no second `-`) AND-s them: "Pods labelled
app=frontend *inside* monitoring" — a different, usually wrong, policy. A bare `podSelector` in
`from` means "in the policy's own namespace".

Every namespace carries the label `kubernetes.io/metadata.name: <name>` automatically, so you
can select a namespace by name without labelling it first.
