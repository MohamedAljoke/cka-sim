# cka-sim docs

How cka-sim works, part by part. Start with the [project README](../README.md) for the overview.

| Doc | What it covers |
| --- | --- |
| [LAB.md](LAB.md) | The lab: Docker in Docker, kind as a Go library, the exam-like base host and ssh, the cluster's state, who can reach what |
| [TERMINAL.md](TERMINAL.md) | The browser terminal: the Connect contract over HTTP/2, the pty, a keystroke end to end, resize, ending a shell |
| [TASKS.md](TASKS.md) | Tasks: the scripts in a task folder, how they run on a node, grading, heal and tidy, the selftest, the exam |
| [HOSTING.md](HOSTING.md) | Labs on demand on Kubernetes: the Machine seam, the lab pod, RBAC and NetworkPolicy, Terraform, kustomize and Caddy |

Development notes are in [dev/](dev/): [STATUS.md](dev/STATUS.md) is what's built,
[BUILD-IN-STEPS.md](dev/BUILD-IN-STEPS.md) what's left, and [ASSISTANT.md](dev/ASSISTANT.md) the plan for the AI
assistant.
