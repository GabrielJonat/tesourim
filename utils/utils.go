package utils

import (
	"math"
	"math/rand"
	"time"
)

// Check if the target node can be reached without visiting any trap nodes
func CanReach(graph map[int][]int, traps map[int]bool, start, target int) bool {
	visited := make(map[int]bool)
	return dfs(graph, traps, visited, start, target)
}

// Depth-First Search function
func dfs(graph map[int][]int, traps map[int]bool, visited map[int]bool, current, target int) bool {
	// If the current node is a trap, return false
	if traps[current] {
		return false
	}

	// If we reach the target, return true
	if current == target {
		return true
	}

	// Mark the current node as visited
	visited[current] = true

	// Explore neighbors
	for _, neighbor := range graph[current] {
		if !visited[neighbor] {
			if dfs(graph, traps, visited, neighbor, target) {
				return true
			}
		}
	}

	return false
}

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// GenerateGraph creates a grid graph with up to 8 connections per node
func GenerateGraph(L int) map[int][]int {
	graph := make(map[int][]int)

	// Calculate total number of nodes
	totalNodes := L * L

	// Helper to get row and column of a node
	getRowCol := func(node int) (int, int) {
		return node / L, node % L
	}

	// Helper to check if a node is within bounds
	isValid := func(row, col int) bool {
		return row >= 0 && row < L && col >= 0 && col < L
	}

	// Directions for the 8 neighbors
	directions := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1}, // Top-left, Top, Top-right
		{0, -1},          {0, 1},  // Left,        Right
		{1, -1}, {1, 0}, {1, 1},   // Bottom-left, Bottom, Bottom-right
	}

	// Build the graph
	for node := 0; node < totalNodes; node++ {
		row, col := getRowCol(node)

		for _, dir := range directions {
			newRow := row + dir[0]
			newCol := col + dir[1]

			if isValid(newRow, newCol) {
				neighbor := newRow*L + newCol
				graph[node] = append(graph[node], neighbor)
			}
		}
	}

	return graph
}

func GenerateTraps(L int, treasure int, dificulty int) (map[int]bool) {

	traps := make(map[int]bool)
	maxNodes := int(L * L)
	maxTraps := int(float64(maxNodes) * 0.75)
	visited := make([]int, 0, maxNodes)
	switch dificulty {
	case 1:
		maxTraps = int(float64(maxTraps) * 0.6)
	case 2:
		maxTraps = int(float64(maxTraps) * 0.8)
	case 3:
		maxTraps = maxTraps
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < int(maxTraps); i++ {
		node := rand.Intn(int(maxNodes))
		visited = append(visited, node)
		for node == treasure || contains(visited, node) {
			node = rand.Intn(int(maxNodes))
		}
		traps[node] = true
	}
	if len(traps) == 0 {
		traps[0] = true
	}
	return traps
}

func GenerateTreasure(L int) int {
	rand.Seed(time.Now().UnixNano())
	treasure := rand.Intn(L*L)
	return treasure
}

func RussianRoulette(dificulty int) bool {
	grandTotal := []int{1, 2, 3, 4, 5, 6}
	randomInt  := rand.Intn(len(grandTotal))
	if dificulty == 3 {
		return grandTotal[randomInt] != 6
	}
	if dificulty == 2 {
		return grandTotal[randomInt] % 2 == 0
	}
	if dificulty == 1 {
		return grandTotal[randomInt] < 3
	}
	return true
}

func RandomFloat64() float64 {
	return rand.Float64()
}

func RandomMoves(currentPos, targetPos float64, gridSize int) float64 {
	// Se estiver próximo do alvo, escolhe um novo alvo
	if math.Abs(currentPos - targetPos) < 0.1 {
		// Retorna um valor entre -1 e 1 para indicar direção do movimento
		return rand.Float64()*2 - 1
	}

	// Move suavemente em direção ao alvo atual
	direction := targetPos - currentPos
	speed := 0.05 // Ajuste este valor para controlar a suavidade do movimento

	// Normaliza o movimento para ser mais suave
	if math.Abs(direction) > speed {
		if direction > 0 {
			return speed
		}
		return -speed
	}

	return direction
}

// calculatePID calcula o movimento baseado no controle PID
func CalculatePID(setpoint float64, currentPos float64, dificulty int, errAcum float64, prevErr float64) float64 {
	var kp, ki, kd float64

	// Ajusta as constantes do PID baseado na dificuldade
	switch dificulty {
	case 1: // Fácil - apenas proporcional
		kp, ki, kd = 0.3, 0, 0
	case 2: // Médio - proporcional + integral
		kp, ki, kd = 0.3, 0.1, 0
	case 3: // Difícil - PID completo
		kp, ki, kd = 0.3, 0.1, 0.05
	}

	err := setpoint - currentPos
	errAcum += err
	derivative := err - prevErr

	output := (kp * err) + (ki * errAcum) + (kd * derivative)
	prevErr = err

	// Limita a saída para movimento suave
	if output > 1 {
		output = 1
	} else if output < -1 {
		output = -1
	}

	return output
}

func CaraOuCoroa() bool {
	r := rand.Intn(2)
	if r == 0 {
		return true
	}
	return false
}