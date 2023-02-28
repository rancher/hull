package parse

import (
	"fmt"
	"text/template/parse"
)

type Node struct {
	node   parse.Node
	parent *Node

	// fieldContext is the root context to use when adding fields
	// For example, in a range or with block, anything that is parsed within it
	// will need to append the fieldContext before adding the field
	//
	// If fieldContext is an empty string, that implies that it was impossible to identify
	// what the root context of this node is, so it's impossible to figure out what
	// a field parsed from within this node is
	//
	// i.e. If you have the following template:
	// {{ range (toYaml .Values.namespaces | fromYaml) }}
	// {{ .data }}
	// {{ end }}
	//
	// (toYaml .Values.namespaces | fromYaml) is considered ambiguous since we don't have
	// a direct reference to a field in the built-in object, so the only thing that
	// should be collected here is .Values.namespace, not .Values.namespace.data, since the
	// only way you could know that (toYaml .Values.namespaces | fromYaml) is equivalent to
	// .Values.namespaces is by understanding how toYaml and fromYaml manipulate the data. This
	// is something that is not capable for Hull to do today.
	fieldContext string
}

func (n *Node) isAmbiguous() bool {
	return len(n.fieldContext) == 0
}

func (n *Node) getFieldContext(fieldContext string) string {
	if n.parent == nil {
		return fieldContext
	}
	if n.parent.isAmbiguous() {
		// ambiguous parent implies ambiguous child; ignore fieldContext provided
		return ""
	}
	if len(fieldContext) == 0 {
		// ambiguous values provided
		return fieldContext
	}
	if fieldContext == "." {
		// use parent fieldContext
		return n.parent.fieldContext
	}
	if n.parent.fieldContext == "." {
		// use child fieldContext
		return fieldContext
	}
	// parent has fieldContext and provided fieldContext is not ambiguous or ., so prepend it
	return n.parent.fieldContext + fieldContext
}

func toNode(node parse.Node, parent *Node, fieldContext string) *Node {
	if node == nil {
		panic(fmt.Errorf("cannot convert nil node toNode"))
	}
	n := &Node{
		node:   node,
		parent: parent,
	}
	n.fieldContext = n.getFieldContext(fieldContext)
	return n
}

func toNodes(parseNodes []parse.Node, parent *Node, fieldContext string) []*Node {
	nodes := make([]*Node, len(parseNodes))
	for i, parseNode := range parseNodes {
		nodes[i] = toNode(parseNode, parent, fieldContext)
	}
	return nodes
}
