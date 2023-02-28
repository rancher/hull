package parse

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
)

type Result struct {
	Fields        []string
	TemplateCalls []string
	EmitWarning   bool
}

func Template(t *template.Template) *Result {
	result := &Result{}
	fields := map[string]bool{}
	templateCalls := map[string]bool{}

	nodes := []*Node{toNode(t.Root, nil, ".")}
	i := 0
	for len(nodes) > i {
		n := nodes[i]
		if n == nil {
			continue
		}
		node := n.node
		switch node := node.(type) {

		// ListNode is the highest level of nodes that starts a tree
		// The child nodes of the ListNode have no correlation to each other; this is where
		// NodeText nodes will exist alongside nodes containing actual evaluation
		case *parse.ListNode:
			// A sequence of nodes, such as everything contained within an if block
			nodes = append(nodes, toNodes(node.Nodes, n, ".")...)

		// NodeIf, NodeRange, NodeWith (BranchNodes)
		// These nodes conditionally have one value or another
		// All follow exactly the same format:
		case *parse.IfNode:
			// {{ if .PipeNode }}
			// {{ .ListNode }}
			// {{ else }}
			// {{ .ElseList }}
			// {{ end }}
			branchNode := node.BranchNode
			nodes = append(nodes, toNode(branchNode.Pipe, n, "."))
			nodes = append(nodes, toNode(branchNode.List, n, "."))
			if branchNode.ElseList != nil {
				nodes = append(nodes, toNode(branchNode.ElseList, n, "."))
			}
		case *parse.RangeNode:
			// {{ range .PipeNode }}
			// {{ .ListNode }}
			// {{ else }}
			// {{ .ElseList }}
			// {{ end }}
			branchNode := node.BranchNode
			nodes = append(nodes, toNode(branchNode.Pipe, n, "."))
			fieldContext, isRoot := getFieldContext(node.Pipe)
			listNode := toNode(branchNode.List, n, fieldContext)
			if isRoot {
				// override fieldContext to be root
				listNode.fieldContext = fieldContext
			}
			nodes = append(nodes, listNode)
			if branchNode.ElseList == nil {
				break
			}
			elseNode := toNode(branchNode.ElseList, n, fieldContext)
			if isRoot {
				// override fieldContext to be root
				elseNode.fieldContext = fieldContext
			}
			nodes = append(nodes, elseNode)
		case *parse.WithNode:
			// {{ with .PipeNode }}
			// {{ .ListNode }}
			// {{ else }}
			// {{ .ElseList }}
			// {{ end }}
			branchNode := node.BranchNode
			nodes = append(nodes, toNode(branchNode.Pipe, n, "."))
			fieldContext, isRoot := getFieldContext(node.Pipe)
			listNode := toNode(branchNode.List, n, fieldContext)
			if isRoot {
				// override fieldContext to be root
				listNode.fieldContext = fieldContext
			}
			nodes = append(nodes, listNode)
			if branchNode.ElseList == nil {
				break
			}
			elseNode := toNode(branchNode.ElseList, n, fieldContext)
			if isRoot {
				// override fieldContext to be root
				elseNode.fieldContext = fieldContext
			}
			nodes = append(nodes, elseNode)

		// NodeAction is a general container node for anything that is within brackets
		// Whatever is in the brackets will be in .Pipe
		case *parse.ActionNode:
			// Either a direct access of a field or anything within parentheses
			// i.e. {{ .Values.data }} -> this is just a NodeAction that contains a NodePipe that contains a NodeField
			// i.e. {{ (.Values.data | toYaml) }} -> this is a NodeAction that contains a NodePipe inside .Pipe
			// is evaluated as a single value.
			nodes = append(nodes, toNode(node.Pipe, n, "."))

		// NodeChain is when the contents within brackets involve using parentheses
		// .Node will be whatever is in the parentheses and .Field will be the chain of accesses after it
		case *parse.ChainNode:
			// i.e. ( .Files.Glob "files/myfile/*" ).AsConfig
			// .Node would be '.Files.Glob "files/myfile/*"'
			// .Field would be [AsConfig]
			nodes = append(nodes, toNode(node.Node, n, "."))
			fieldContext, isRoot := getFieldContext(node)
			if len(fieldContext) == 0 {
				break
			}
			if isRoot {
				fields[fieldContext] = true
			} else {
				// TODO: look into
				fields[n.getFieldContext(fieldContext)] = true
			}

		// NodePipe is used for two things: declaring and instantiating variable values
		// and evaluating pipelines of values that result in a single evaluated value
		case *parse.PipeNode:
			// i.e. $_ := .Values.labels
			// Variable declarations are in .Decl
			// Anything after the declaration is in .Cmds, where each Cmd is separated by a `|`
			// i.e. .Values.labels | toYaml -> .Cmds = [ .Values.labels toYaml ]
			// i.e. fromYaml .Values.labels | toYaml -> .Cmds = [ NodeCommand(fromYaml .Values.labels) toYaml ]
			for _, cmd := range node.Cmds {
				nodes = append(nodes, toNode(cmd, n, "."))
			}
			// Note: .Decl is ignored since there's no fields in it

		// NodeCommand is a single command that needs to be evaluated
		case *parse.CommandNode:
			// i.e. toYaml .Values.data
			if !isIncludeCommand(node) {
				nodes = append(nodes, toNodes(node.Args, n, ".")...)
				break
			}
			// special logic to handle 'include' block
			if node.Args[1].Type() == parse.NodeString {
				// Add the string contents as a template that has been called
				stringNode := node.Args[1].(*parse.StringNode)
				templateCalls[stringNode.Text] = true
			} else {
				// This is another thing to be evaluated; add it back to the stack
				// No need to add it as a template call since it's unclear what template would be called
				nodes = append(nodes, toNode(node.Args[1], n, "."))
			}
			if !isDotNode(node.Args[2]) {
				// If the argument to includes is a '.', we can safely ignore it
				// since it's not intended to be used as global logic.
				// But if it's not a '.', we need to evaluate it.
				nodes = append(nodes, toNode(node.Args[2], n, "."))
			}

		case *parse.TemplateNode:
			// An action to invoke a template
			// {{ template .Name .Pipe }}
			templateCalls[node.Name] = true
			if !isDotNode(node.Pipe) {
				// evaluate only if it is not the global '.'
				nodes = append(nodes, toNode(node.Pipe, n, "."))
			}

		// NodeDot is just a .
		case *parse.DotNode:
			// If this is ever parsed, we need to indicate that this file may be doing
			// something complex that cannot be captured via coverage. This is why we
			// mark it as using the built in object directly
			if n.fieldContext == "." {
				// we are only using the built in object if the fieldContext is a .
				result.EmitWarning = true
			}

		// NodeField is any access of a field within the built-in object passed in to the template
		case *parse.FieldNode:
			// i.e. {{ .Values.data }}
			// i.e. {{ .Chart.Name }}
			// i.e. {{ .Capabilities.KubeVersion }}
			fields[n.getFieldContext(node.String())] = true

		// NodeVariable is any variable
		// Note: We only really care about $, the root variable
		case *parse.VariableNode:
			// A variable
			// i.e. $hello
			// i.e. $ (the root variable)
			if !isRootVariable(node) {
				break
			}
			if len(node.Ident) <= 1 {
				break
			}
			// A call that looks like $.Values.* or $.Capabilities.*
			rootField := "." + strings.Join(node.Ident[1:], ".")
			fields[rootField] = true

		case *parse.IdentifierNode, *parse.TextNode, *parse.BoolNode, *parse.NilNode, *parse.NumberNode, *parse.StringNode, *parse.CommentNode, *parse.BreakNode, *parse.ContinueNode:
			// do nothing; these are irrelevant for coverage
		default:
			panic(fmt.Errorf("unknown node type: %t", node))
		}
		i++
	}

	// get final values
	delete(fields, "")
	delete(templateCalls, "")
	for field := range fields {
		result.Fields = append(result.Fields, field)
	}
	sort.Strings(result.Fields)
	for templateCall := range templateCalls {
		result.TemplateCalls = append(result.TemplateCalls, templateCall)
	}
	sort.Strings(result.TemplateCalls)
	return result
}
