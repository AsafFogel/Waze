// cmd/filter_graph/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Node struct {
	Id   int     `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type Edge struct {
	Id         int     `json:"id"`
	From       int     `json:"from"`
	To         int     `json:"to"`
	Length     float64 `json:"length"`
	SpeedLimit float64 `json:"speedlimit"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input_file> <output_file>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// טעינת הגרף
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var graph GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Printf("Loaded: %d nodes, %d edges\n", len(graph.Nodes), len(graph.Edges))


	// בניית רשימות שכנויות
	adjList := make(map[int][]int)
	reverseAdj := make(map[int][]int)

	for _, edge := range graph.Edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
		reverseAdj[edge.To] = append(reverseAdj[edge.To], edge.From)
	}

	// איסוף כל הצמתים
	nodeSet := make(map[int]bool)
	for _, node := range graph.Nodes {
		nodeSet[node.Id] = true
	}

	// שלב 1: סריקה ראשונה לפי סדר סיום
	visited := make(map[int]bool)
	finishOrder := make([]int, 0)

	var dfs1 func(node int)
	dfs1 = func(node int) {
		visited[node] = true
		for _, neighbor := range adjList[node] {
			if !visited[neighbor] {
				dfs1(neighbor)
			}
		}
		finishOrder = append(finishOrder, node)
	}

	for nodeId := range nodeSet {
		if !visited[nodeId] {
			dfs1(nodeId)
		}
	}

	// שלב 2: סריקה בגרף ההפוך לפי סדר סיום הפוך
	visited = make(map[int]bool)
	var components [][]int

	var dfs2 func(node int, component *[]int)
	dfs2 = func(node int, component *[]int) {
		visited[node] = true
		*component = append(*component, node)
		for _, neighbor := range reverseAdj[node] {
			if !visited[neighbor] {
				dfs2(neighbor, component)
			}
		}
	}

	for i := len(finishOrder) - 1; i >= 0; i-- {
		node := finishOrder[i]
		if !visited[node] {
			component := make([]int, 0)
			dfs2(node, &component)
			components = append(components, component)
		}
	}

	// מציאת הרכיב הגדול ביותר
	var largestComponent []int
	for _, comp := range components {
		if len(comp) > len(largestComponent) {
			largestComponent = comp
		}
	}

	fmt.Printf("Found %d strongly connected components\n", len(components))

	fmt.Printf("Largest component: %d nodes\n", len(largestComponent))

	// סינון הגרף
	keepNodes := make(map[int]bool)
	for _, nodeId := range largestComponent {
		keepNodes[nodeId] = true
	}

	var filteredNodes []Node
	for _, node := range graph.Nodes {
		if keepNodes[node.Id] {
			filteredNodes = append(filteredNodes, node)
		}
	}

	var filteredEdges []Edge
	for _, edge := range graph.Edges {
		if keepNodes[edge.From] && keepNodes[edge.To] {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	fmt.Printf("Post Filtering: %d nodes, %d edges\n", len(filteredNodes), len(filteredEdges))

	// שמירה
	filtered := GraphData{
		Nodes: filteredNodes,
		Edges: filteredEdges,
	}

	output, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		fmt.Printf("Error creating the file: %v\n", err)
		return
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		fmt.Printf("Error Saving the file: %v\n", err)
		return
	}

	fmt.Printf("Saved to: %s\n", outputFile)
}