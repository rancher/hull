package main

name = input.metadata.name

tolerates_windows {
	has_windows_node_selectors
	has_windows_tolerations
}

has_windows_node_selectors {
	input.spec.template.spec.nodeSelector["kubernetes.io/os"] == "linux"
}

has_windows_tolerations {
	toleration := input.spec.template.spec.tolerations[_]
	toleration.key == "cattle.io/os"
	toleration.value == "linux"
	toleration.effect == "NoSchedule"
	toleration.operator == "Equal"
}

deny[msg] {
	input.kind = "Deployment"
	not tolerates_windows
	msg = sprintf("Pod %s does not tolerate windows", [name])
}