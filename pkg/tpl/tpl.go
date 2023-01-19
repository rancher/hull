package tpl

import (
	"text/template"
	"text/template/parse"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/sirupsen/logrus"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

func getDummyFuncMap() template.FuncMap {
	funcMap := funcMap()
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		return "", nil
	}
	funcMap["tpl"] = func(tpl string, vals helmChartUtil.Values) (string, error) {
		return "", nil
	}
	funcMap["required"] = func(warn string, val interface{}) (interface{}, error) {
		return val, nil
	}
	funcMap["fail"] = func(msg string) (string, error) {
		return "", nil
	}
	funcMap["lookup"] = func(apiversion string, resource string, namespace string, name string) (map[string]interface{}, error) {
		return nil, nil
	}
	return funcMap
}

func CollectBuiltInObjectsUsage(chartPath string) (map[string]bool, error) {
	var usage map[string]bool

	c, err := chart.NewChart(chartPath)
	if err != nil {
		return nil, err
	}

	ch := c.GetHelmChart()
	for _, tpl := range ch.Templates {
		if tpl == nil {
			continue
		}
		t, err := template.New(tpl.Name).Funcs(getDummyFuncMap()).Parse(string(tpl.Data))
		if err != nil {
			return nil, err
		}
		nodes := []parse.Node{t.Root}
		i := 0
		for len(nodes) > i {
			node := nodes[i]
			if node == nil {
				continue
			}
			switch node.Type() {
			case parse.NodeText:
				// do nothing
			case parse.NodeList:
				listNode := node.(*parse.ListNode)
				nodes = append(nodes, listNode.Nodes...)
			case parse.NodeAction:
				actionNode := node.(*parse.ActionNode)
				nodes = append(nodes, actionNode.Pipe)
			case parse.NodePipe:
				pipeNode := node.(*parse.PipeNode)
				for _, cmd := range pipeNode.Cmds {
					nodes = append(nodes, cmd)
				}
			case parse.NodeIf:
				branchNode := node.(*parse.IfNode).BranchNode
				nodes = append(nodes, branchNode.Pipe, branchNode.List)
				if branchNode.ElseList != nil {
					nodes = append(nodes, branchNode.ElseList)
				}
			case parse.NodeRange:
				branchNode := node.(*parse.RangeNode).BranchNode
				nodes = append(nodes, branchNode.Pipe, branchNode.List)
				if branchNode.ElseList != nil {
					nodes = append(nodes, branchNode.ElseList)
				}
			case parse.NodeWith:
				branchNode := node.(*parse.WithNode).BranchNode
				nodes = append(nodes, branchNode.Pipe, branchNode.List)
				if branchNode.ElseList != nil {
					nodes = append(nodes, branchNode.ElseList)
				}
			case parse.NodeCommand:
				commandNode := node.(*parse.CommandNode)
				nodes = append(nodes, commandNode.Args...)
			// case parse.NodeChain:
			// 	chainNode := node.(*parse.ChainNode)
			// 	logrus.Infof("chainNode: %s, %s", chainNode.Field, chainNode.Node)
			// case parse.NodeDot:
			// case parse.NodeField:
			// case parse.NodeIdentifier:
			// case parse.NodeTemplate:
			default:
				logrus.Infof("%d: %v", node.Type(), node)
			}
			i++
		}
	}

	return usage, nil
}
