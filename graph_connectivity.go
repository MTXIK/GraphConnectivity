package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
)

// Graph представляет неориентированный граф с использованием списка смежности
type Graph struct {
	Vertices int     // Количество вершин в графе
	AdjList  [][]int // Список смежности, где AdjList[i] содержит список вершин, смежных с вершиной i
}

// NewGraph создаёт новый граф с заданным числом вершин
func NewGraph(vertices int) *Graph {
	adj := make([][]int, vertices) // Инициализируем список смежности для каждой вершины
	for i := range adj {
		adj[i] = []int{} // Для каждой вершины создаём пустой срез для соседей
	}
	return &Graph{
		Vertices: vertices, // Устанавливаем количество вершин
		AdjList:  adj,      // Присваиваем список смежности
	}
}

// AddEdge добавляет неориентированное ребро между вершинами u и v
func (g *Graph) AddEdge(u, v int) {
	g.AdjList[u] = append(g.AdjList[u], v) // Добавляем v в список смежности u
	g.AdjList[v] = append(g.AdjList[v], u) // Добавляем u в список смежности v (так как граф неориентированный)
}

// ReadGraph считывает граф из бинарного файла
// Формат файла: int16 размер, затем матрица смежности размер x размер, элементы типа int16
func ReadGraph(filename string) (*Graph, error) {
	file, err := os.Open(filename) // Открываем бинарный файл для чтения
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии входного файла: %v", err)
	}
	defer file.Close() // Гарантируем закрытие файла после завершения функции

	var size int16
	// Считываем размер графа (количество вершин)
	err = binary.Read(file, binary.LittleEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении размера графа: %v", err)
	}

	if size <= 0 {
		return nil, fmt.Errorf("недопустимый размер графа: %d", size)
	}

	// Считываем матрицу смежности
	adjMatrix := make([][]int, size)
	for i := int16(0); i < size; i++ {
		adjMatrix[i] = make([]int, size)
		for j := int16(0); j < size; j++ {
			var val int16
			err = binary.Read(file, binary.LittleEndian, &val)
			if err != nil {
				return nil, fmt.Errorf("ошибка при чтении матрицы смежности на позиции (%d,%d): %v", i, j, err)
			}
			adjMatrix[i][j] = int(val)
		}
	}

	// Преобразуем матрицу смежности в список смежности для удобства обработки
	graph := NewGraph(int(size))
	for i := 0; i < int(size); i++ {
		for j := 0; j < int(size); j++ {
			if adjMatrix[i][j] != 0 && i < j { // Проверяем наличие ребра и избегаем дублирования (так как граф неориентированный)
				graph.AddEdge(i, j)
			}
		}
	}

	return graph, nil
}

//Алг. Тарьяна
// ArticulationPointsAndBridges находит точки сочленения и мосты в графе
func (g *Graph) ArticulationPointsAndBridges() (articulationPoints []int, bridges [][2]int) {
	visited := make([]bool, g.Vertices)    // Массив для отслеживания посещённых вершин
	discovery := make([]int, g.Vertices)   // Время обнаружения каждой вершины
	low := make([]int, g.Vertices)         // Низшее время, доступное из поддерева вершины
	parent := make([]int, g.Vertices)      // Родитель каждой вершины в DFS-дереве
	for i := range parent {
		parent[i] = -1 // Инициализируем родителя как -1 (нет родителя)
	}
	time := 0 // Глобальное время для DFS
	ap := make(map[int]bool) // Множество точек сочленения
	br := [][2]int{}          // Список мостов

	// Рекурсивная функция DFS для обхода графа и вычисления low
	var dfs func(u int)
	dfs = func(u int) {
		visited[u] = true                // Отмечаем вершину как посещённую
		discovery[u] = time              // Устанавливаем время обнаружения вершины u
		low[u] = time                     // Инициализируем low[u] текущим временем
		time++                            // Увеличиваем глобальное время
		children := 0                      // Количество дочерних вершин в DFS-дереве

		for _, v := range g.AdjList[u] { // Проходим по всем смежным вершинам v вершины u
			if !visited[v] { // Если вершина v ещё не посещена
				children++            // Увеличиваем счётчик дочерних вершин
				parent[v] = u         // Устанавливаем u как родителя для v
				dfs(v)                // Рекурсивно вызываем DFS для вершины v

				low[u] = min(low[u], low[v]) // Обновляем low[u] как минимум из текущего low[u] и low[v]

				// Проверяем, является ли вершина u точкой сочленения

				// Условие 1: Если u - корень DFS и имеет более одного дочернего поддерева
				if parent[u] == -1 && children > 1 {
					ap[u] = true // Вершина u является точкой сочленения
				}
				//Если у корня DFS дерева более одного дочернего поддерева, это означает, 
				//что существует более одной подгруппы вершин, которые связаны через корень, но не связаны между собой напрямую.
				//То есть, все эти поддеревья соединены только через корень.

				// Условие 2: Если u не корень и low[v] >= discovery[u]
				if parent[u] != -1 && low[v] >= discovery[u] {
					ap[u] = true // Вершина u является точкой сочленения
				}
				
				// Точки сочленения: Вершина u является точкой сочленения, если существует хотя бы один потомок v в дереве DFS такой, 
				// что low[v] >= discovery[u]. Это означает, что нет обратного пути из поддерева v, который мог бы вернуться к предкам u, кроме самого u. 
				// Удаление u разрывает граф на отдельные компоненты.

				if low[v] > discovery[u] {
					br = append(br, [2]int{u, v}) // Ребро (u, v) добавляется в список мостов
				}
				
				// Мосты: Ребро (u, v) является мостом, если low[v] > discovery[u]. 
				// Это означает, что нет других путей из поддерева v, которые могли бы соединиться с предками u, кроме через ребро (u, v). 
				// Удаление такого ребра увеличивает количество компонент связности.
				
			} else if v != parent[u] { // Если вершина v уже посещена и не является родителем u
				// Обновляем low[u] как минимум из текущего low[u] и discovery[v]
				// Это учитывает обратное ребро (back edge) от u к v
				low[u] = min(low[u], discovery[v])
				//Уменьшение low[u] происходит:
				//Через обратные рёбра: Когда из вершины u существует обратный путь к более ранней вершине v.
				//Через дочерние вершины: Если из поддерева дочерней вершины v существует путь к более ранней вершине, чем текущая вершина u.
			}
		}
	}

	// Запускаем DFS для всех компонент связности графа
	for u := 0; u < g.Vertices; u++ {
		if !visited[u] {
			dfs(u) // Запускаем DFS для непосещённой вершины u
		}
	}

	// Собираем все точки сочленения из карты в срез
	for k := range ap {
		articulationPoints = append(articulationPoints, k)
	}

	return articulationPoints, br // Возвращаем список точек сочленения и мостов
}

//Алг. Тарьяна
// BiconnectedComponents находит компоненты двусвязности графа
func (g *Graph) BiconnectedComponents() [][][2]int {
	visited := make([]bool, g.Vertices)    // Массив для отслеживания посещённых вершин
	discovery := make([]int, g.Vertices)   // Время обнаружения каждой вершины
	low := make([]int, g.Vertices)         // Низшее время, доступное из поддерева вершины
	parent := make([]int, g.Vertices)      // Родитель каждой вершины в DFS-дереве
	for i := range parent {
		parent[i] = -1 // Инициализируем родителя как -1 (нет родителя)
	}
	time := 0                        // Глобальное время для DFS
	stack := [][2]int{}              // Стек для хранения рёбер текущей компоненты
	bcc := [][][2]int{}              // Список всех компонент двусвязности

	// Рекурсивная функция DFS для обхода графа и вычисления low
	var dfs func(u int)
	dfs = func(u int) {
		visited[u] = true                // Отмечаем вершину как посещённую
		discovery[u] = time              // Устанавливаем время обнаружения вершины u
		low[u] = time                     // Инициализируем low[u] текущим временем
		time++                            // Увеличиваем глобальное время

		for _, v := range g.AdjList[u] { // Проходим по всем смежным вершинам v вершины u
			if !visited[v] { // Если вершина v ещё не посещена
				parent[v] = u         // Устанавливаем u как родителя для v
				stack = append(stack, [2]int{u, v}) // Добавляем ребро (u, v) в стек
				dfs(v)                // Рекурсивно вызываем DFS для вершины v

				low[u] = min(low[u], low[v]) // Обновляем low[u] как минимум из текущего low[u] и low[v]

				// Проверяем, разделяет ли ребро (u, v) компоненты двусвязности
					// Этот участок кода отвечает за выделение новой компоненты двусвязности после обнаружения условия, 
					// при котором текущая вершина u разделяет граф на компоненты двусвязности. 
					// Основная задача — извлечь все рёбра, принадлежащие этой новой компоненте, 
					// из стека и сохранить их в списке компонент bcc.
					//
					// проверка условия low[v] >= discovery[u] позволяет определить, 
					// когда текущая вершина u разделяет граф на независимые компоненты
				if low[v] >= discovery[u] {
					component := [][2]int{} // Создаём новую компоненту двусвязности
					for {
						if len(stack) == 0 {
							break // Если стек пуст, выходим из цикла
						}
						edge := stack[len(stack)-1] // Берём последнее ребро из стека
						stack = stack[:len(stack)-1] // Удаляем это ребро из стека
						component = append(component, edge) // Добавляем ребро в текущую компоненту
						if edge[0] == u && edge[1] == v { // Если достигли разделяющего ребра
							break // Завершаем сбор текущей компоненты
						}
					}
					bcc = append(bcc, component) // Добавляем компоненту в список компонент двусвязности
				}
			} else if v != parent[u] && discovery[v] < discovery[u] { // Если вершина v уже посещена, не является родителем, и была обнаружена раньше
				low[u] = min(low[u], discovery[v]) // Обновляем low[u] как минимум из текущего low[u] и discovery[v]
				stack = append(stack, [2]int{u, v}) // Добавляем ребро (u, v) в стек
				//Уменьшение low[u] происходит:
				//Через обратные рёбра: Когда из вершины u существует обратный путь к более ранней вершине v.
				//Через дочерние вершины: Если из поддерева дочерней вершины v существует путь к более ранней вершине, чем текущая вершина u.
			}
		}
	}

	// Запускаем DFS для всех компонент связности графа
	for u := 0; u < g.Vertices; u++ {
		if !visited[u] {
			dfs(u) // Запускаем DFS для непосещённой вершины u
			// После завершения DFS, если в стеке остались рёбра, они образуют отдельную компоненту двусвязности
			if len(stack) > 0 {
				bcc = append(bcc, stack) // Добавляем оставшиеся рёбра как отдельную компоненту
				stack = [][2]int{}        // Очищаем стек для следующей компоненты
			}
		}
	}

	return bcc // Возвращаем список компонент двусвязности
}

// ConnectedComponents находит компоненты связности графа
func (g *Graph) ConnectedComponents() [][]int {
	visited := make([]bool, g.Vertices) // Массив для отслеживания посещённых вершин
	components := [][]int{}              // Список всех компонент связности

	// Рекурсивная функция DFS для поиска компонент связности
	var dfs func(u int, component *[]int)
	dfs = func(u int, component *[]int) {
		visited[u] = true                    // Отмечаем вершину как посещённую
		*component = append(*component, u)   // Добавляем вершину в текущую компоненту
		for _, v := range g.AdjList[u] {    // Проходим по всем смежным вершинам v вершины u
			if !visited[v] {
				dfs(v, component) // Рекурсивно вызываем DFS для непосещённой вершины v
			}
		}
	}

	// Запускаем DFS для всех компонент связности графа
	for u := 0; u < g.Vertices; u++ {
		if !visited[u] {
			component := []int{}   // Создаём новую компоненту связности
			dfs(u, &component)     // Запускаем DFS для вершины u
			components = append(components, component) // Добавляем компоненту в список
		}
	}

	return components // Возвращаем список компонент связности
}

// min возвращает минимальное из двух целых чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Определение и разбор командных флагов
	outputFile := flag.String("o", "output.txt", "Имя выходного файла") // Флаг -o для задания имени выходного файла
	flag.Parse() // Разбираем флаги

	// Проверяем наличие обязательного аргумента (имени входного файла)
	if flag.NArg() < 1 {
		log.Fatalf("Использование: %s inputfile [-o outputfile]", os.Args[0])
	}

	inputFile := flag.Arg(0) // Получаем имя входного файла из аргументов
	outFile := *outputFile   // Имя выходного файла (либо по умолчанию, либо задано через флаг)

	// Считываем граф из бинарного файла
	graph, err := ReadGraph(inputFile)
	if err != nil {
		log.Fatalf("Не удалось считать граф: %v", err)
	}

	// Находим точки сочленения и мосты
	articulationPoints, bridges := graph.ArticulationPointsAndBridges()

	// Находим компоненты двусвязности
	bcc := graph.BiconnectedComponents()

	// Находим компоненты связности
	cc := graph.ConnectedComponents()

	// Открываем (или создаём) выходной файл для записи результатов
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Ошибка при создании выходного файла: %v", err)
	}
	defer f.Close() // Гарантируем закрытие файла после завершения функции

	// Записываем раздел a) Мосты и точки сочленения
	_, err = f.WriteString("a) Мосты и точки сочленения:\n")
	if err != nil {
		log.Fatalf("Ошибка при записи в выходной файл: %v", err)
	}

	// Записываем точки сочленения
	_, err = f.WriteString("Точки сочленения:\n")
	if len(articulationPoints) == 0 {
		_, _ = f.WriteString("Отсутствуют\n")
	} else {
		for _, ap := range articulationPoints {
			_, _ = f.WriteString(fmt.Sprintf("%d ", ap))
		}
		_, _ = f.WriteString("\n")
	}

	// Записываем мосты
	_, err = f.WriteString("Мосты:\n")
	if len(bridges) == 0 {
		_, _ = f.WriteString("Отсутствуют\n")
	} else {
		for _, bridge := range bridges {
			_, _ = f.WriteString(fmt.Sprintf("(%d, %d) ", bridge[0], bridge[1]))
		}
		_, _ = f.WriteString("\n")
	}

	// Записываем раздел b) Компоненты двусвязности
	_, err = f.WriteString("\nb) Компоненты двусвязности:\n")
	if err != nil {
		log.Fatalf("Ошибка при записи в выходной файл: %v", err)
	}
	for i, component := range bcc {
		_, err = f.WriteString(fmt.Sprintf("Компонента %d:\n", i+1))
		if err != nil {
			log.Fatalf("Ошибка при записи в выходной файл: %v", err)
		}
		for _, edge := range component {
			_, err = f.WriteString(fmt.Sprintf("(%d, %d) ", edge[0], edge[1]))
			if err != nil {
				log.Fatalf("Ошибка при записи в выходной файл: %v", err)
			}
		}
		_, err = f.WriteString("\n")
		if err != nil {
			log.Fatalf("Ошибка при записи в выходной файл: %v", err)
		}
	}

	// Записываем раздел c) Компоненты связности
	_, err = f.WriteString("\nc) Компоненты связности:\n")
	if err != nil {
		log.Fatalf("Ошибка при записи в выходной файл: %v", err)
	}
	for i, component := range cc {
		_, err = f.WriteString(fmt.Sprintf("Компонента %d: ", i+1))
		if err != nil {
			log.Fatalf("Ошибка при записи в выходной файл: %v", err)
		}
		for _, vertex := range component {
			_, err = f.WriteString(fmt.Sprintf("%d ", vertex))
			if err != nil {
				log.Fatalf("Ошибка при записи в выходной файл: %v", err)
			}
		}
		_, err = f.WriteString("\n")
		if err != nil {
			log.Fatalf("Ошибка при записи в выходной файл: %v", err)
		}
	}

	// Выводим сообщение об успешном завершении
	fmt.Printf("Анализ завершён. Результаты записаны в %s\n", outFile)
}
