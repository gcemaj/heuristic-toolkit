package toolkit

import (
	"fmt"
	"sync"
	"time"
)

type Step struct {
	From      int
	To        int
	StartTime int
	EndTime   int
}

type Solution struct {
	Score    float64
	Solution [][]Step
}

type AntColonyOptimization struct {
	// MaxTimeout is the maximum time in seconds that the optimization has to run
	MaxTimeout   int
	Ants         []Ant
	Graph        Graph
	BestSolution Solution
	Solutions    []*Solution
	Alpha        float64
	Beta         float64
}

func NewAntColonyOptimization(maxTimeout int, ants []Ant, alpha, beta float64, graph Graph) AntColonyOptimization {
	return AntColonyOptimization{
		MaxTimeout: maxTimeout,
		Ants:       ants,
		Graph:      graph,
		Alpha:      alpha,
		Beta:       beta,
	}
}

func (ac *AntColonyOptimization) runAntParallel(antId int, ant Ant, wg *sync.WaitGroup) {
	defer wg.Done()
	ant.Reset()
	solution := ant.ComputeSolution(ac.Alpha, ac.Beta, &ac.Graph)
	score := ant.ComputeFitness()

	// deacrease phermones for ant solution
	for _, path := range solution {
		for _, step := range path {
			if step.To >= 0 {
				ac.Graph.Edges[step.From][step.To].Phermone *= 0.7
			}

			if step.From >= 0 {
				ac.Graph.Edges[step.To][step.From].Phermone *= 0.7
			}
		}
	}

	ac.Solutions[antId] = &Solution{
		Score:    score,
		Solution: solution,
	}
}

func (ac *AntColonyOptimization) runOptimization(channel chan int) {
	for {
		var wg sync.WaitGroup

		ac.Solutions = make([]*Solution, len(ac.Ants))

		for i, ant := range ac.Ants {
			wg.Add(1)
			go ac.runAntParallel(i, ant, &wg)
		}

		wg.Wait()

		for _, solution := range ac.Solutions {
			if solution.Score > ac.BestSolution.Score {
				ac.BestSolution.Score = solution.Score
				ac.BestSolution.Solution = solution.Solution
			}
		}

		// increase phermones for best score
		for _, path := range ac.BestSolution.Solution {
			for _, step := range path {
				if step.To >= 0 {
					ac.Graph.Edges[step.From][step.To].Phermone *= 1.4
				}

				if step.From >= 0 {
					ac.Graph.Edges[step.To][step.From].Phermone *= 1.4
				}
			}
		}

		// evaporation
		for _, tos := range ac.Graph.Edges {
			for _, edge := range tos {
				edge.Phermone *= 0.6
			}
		}
	}
}

func (ac *AntColonyOptimization) Run() *Solution {

	optimizationChannel := make(chan int, 1)

	go ac.runOptimization(optimizationChannel)

	select {
	case res := <-optimizationChannel:
		fmt.Println(res)
	case <-time.After(time.Duration(ac.MaxTimeout) * time.Second):
		return &ac.BestSolution
	}

	return nil
}
