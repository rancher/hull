package parse

import (
	"fmt"
	"strings"
	"text/template/parse"
)

func isIncludeCommand(commandNode *parse.CommandNode) bool {
	if len(commandNode.Args) != 3 {
		return false
	}
	node := commandNode.Args[0]
	if node.Type() != parse.NodeIdentifier {
		return false
	}
	identifierNode := node.(*parse.IdentifierNode)
	return identifierNode.Ident == "include"
}

func isRootVariable(variableNode *parse.VariableNode) bool {
	if len(variableNode.Ident) == 0 {
		panic(fmt.Errorf("invalid variable node with no identity"))
	}
	return variableNode.Ident[0] == "$"
}

func isDotNode(node parse.Node) bool {
	switch node := node.(type) {
	case *parse.DotNode:
		return true
	case *parse.PipeNode:
		if len(node.Cmds) != 1 {
			return false
		}
		if len(node.Cmds[0].Args) != 1 {
			return false
		}
		return isDotNode(node.Cmds[0].Args[0])
	default:
		return false
	}
}

func getFieldContext(node parse.Node) (string, bool) {
	switch node := node.(type) {
	// NOTE: These cases have been removed since it's not possible for getFieldContext
	// today to be called on these types, since getFieldContext only gets called for
	// BranchNode or ChainNodes and those cannot contain ActionNodes (brackets within brackets)
	// or ListNodes (lists of items wouldn't belong in the middle of a chain)
	//
	// TODO: Figure out if these need to be added back inside when an example is found
	//
	// case *parse.ListNode:
	// 	// unfurl the contents if there's just one item in the list, if not it is ambiguous
	// 	if len(node.Nodes) > 1 || len(node.Nodes) == 0 {
	// 		return "", false
	// 	}
	// 	return getFieldContext(node.Nodes[0])
	// case *parse.ActionNode:
	// 	// whatever is inside the action
	// 	return getFieldContext(node.Pipe)
	case *parse.ChainNode:
		fieldContext, isRootContext := getFieldContext(node.Node)
		if len(fieldContext) == 0 {
			// ambiguous context was found
			return "", false
		}
		// i.e. (.Values).data.world would return .Values.data.world
		return fieldContext + "." + strings.Join(node.Field, "."), isRootContext
	case *parse.PipeNode:
		if len(node.Decl) > 0 {
			return "", false
		}
		if len(node.Cmds) == 0 || len(node.Cmds) > 1 {
			// cannot parse pipeline into unambiguous field
			// i.e. .Values.data | toYaml
			return "", false
		}
		return getFieldContext(node.Cmds[0])
	case *parse.CommandNode:
		if len(node.Args) == 0 || len(node.Args) > 1 {
			// cannot parse multiple commands into unambiguous field
			// i.e. toYaml .Values.data
			return "", false
		}
		return getFieldContext(node.Args[0])

	case *parse.DotNode:
		// fieldContext is a .
		return ".", false

	case *parse.FieldNode:
		// specific field is found
		return node.String(), false

	case *parse.VariableNode:
		if !isRootVariable(node) {
			// all variables other than root are ambiguous
			return "", false
		}
		if len(node.Ident) == 1 {
			// should be empty $ variable, which is the root of the built-in object
			return ".", true
		}
		// A call that looks like $.Values.* or $.Capabilities.*
		return "." + strings.Join(node.Ident[1:], "."), true
	default:
		panic(fmt.Errorf("cannot getFieldContext for node of type %t", node))
	}
}
