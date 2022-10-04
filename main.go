package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gcemaj/heuristic-toolkit/toolkit"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseFile(fileName string) ([]toolkit.Patient, []int) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil, nil
	}

	patients := []toolkit.Patient{}
	ambulances := []int{}
	lines := strings.Split(string(contents), "\n")
	patientMode := true

	i := 0

	for _, line := range lines[1:] {

		i++

		if line == "" {
			i += 2
			break
		}

		if patientMode {
			vals := strings.Split(line, ",")
			street, _ := strconv.Atoi(vals[0])
			ave, _ := strconv.Atoi(vals[1])
			timeToLive, _ := strconv.Atoi(vals[2])

			patients = append(patients, toolkit.Patient{
				Loc: toolkit.Location{
					Street: street,
					Avenue: ave,
				},
				TimeToLive: timeToLive,
			})
		}
	}

	for _, line := range lines[i:] {
		numAmbulances, _ := strconv.Atoi(line)
		ambulances = append(ambulances, numAmbulances)
	}

	return patients, ambulances
}

func getHospitalLocations(patients []toolkit.Patient, numAmbulances []int) []toolkit.Location {
	km, _ := kmeans.NewWithOptions(0.01, nil)

	coords := []clusters.Observation{}

	for _, i := range patients {
		coords = append(coords, i.Loc)
	}
	centers, _ := km.Partition(coords, len(numAmbulances))

	sortedAmbulances := []int{}
	copyAmbulances := []int{}
	for _, v := range numAmbulances {
		sortedAmbulances = append(sortedAmbulances, v)
		copyAmbulances = append(copyAmbulances, v)
	}
	sort.Slice(sortedAmbulances, func(i, j int) bool {
		return sortedAmbulances[i] > sortedAmbulances[j]
	})

	sortedMap := []int{}

	for _, v1 := range sortedAmbulances {
		for j, v2 := range copyAmbulances {
			if v1 == v2 {
				sortedMap = append(sortedMap, j)
				copyAmbulances[j] = -1
				break
			}
		}
	}

	hospitalLocs := make([]toolkit.Location, len(numAmbulances))

	sort.Slice(centers, func(i, j int) bool {
		return len(centers[i].Observations) > len(centers[j].Observations)
	})

	for i, c := range centers {

		hospitalLocs[sortedMap[i]] = toolkit.Location{
			Street: int(c.Center[0]),
			Avenue: int(c.Center[1]),
		}
		// hospitalLocs = append(hospitalLocs, toolkit.Location{
		// 	Street: int(c.Center[0]),
		// 	Avenue: int(c.Center[1]),
		// })
	}

	return hospitalLocs
}

func main() {

	numAnts := 50

	maxTime := -1
	// patients, numAmbulances := parseFile("inputs/ambulance-input.txt")
	patients, numAmbulances := parseFile("input_data.txt")
	hospitals := getHospitalLocations(patients, numAmbulances)

	graph := toolkit.NewGraph()
	for i, p1 := range patients {
		maxTime = max(maxTime, p1.TimeToLive)
		for j, p2 := range patients {
			if i == j {
				continue
			}
			graph.UpsertEdge(i, j, &toolkit.Edge{
				Phermone: 1.0,
				Distance: toolkit.ComputeDistance(p1.Loc, p2.Loc),
			})
		}

		for hi, h := range hospitals {
			graph.UpsertEdge(-hi-1, i, &toolkit.Edge{
				Phermone: 1.0,
				Distance: toolkit.ComputeDistance(p1.Loc, h),
			})
		}
	}

	ambulances := []toolkit.Ambulance{}
	for i, n := range numAmbulances {
		for j := 0; j < n; j++ {
			ambulances = append(ambulances, toolkit.Ambulance{
				CurrentLocation:  -i - 1,
				StartingLocation: -i - 1,
				CurrentTime:      0,
				Patients:         []int{},
			})
		}
	}

	ants := []toolkit.Ant{}
	for i := 0; i < numAnts; i++ {
		ants = append(ants, &toolkit.HospitalAnt{
			Patients:   patients,
			Hospitals:  hospitals,
			Visited:    make([]bool, len(patients)),
			Ambulances: ambulances,
			Saved:      make([]int, len(patients)),
			MaxTime:    maxTime,
		})
	}

	ac := toolkit.NewAntColonyOptimization(110, ants, 1, 3, graph)

	bestSolution := ac.Run()

	fmt.Println(bestSolution.Score)
	fmt.Println()

	f, _ := os.Create("Outputs/checkmate.txt")

	for i, h := range hospitals {
		f.WriteString(fmt.Sprintf("H%d:%d,%d\n", i+1, h.Street, h.Avenue))
	}

	f.WriteString("\n")

	for len(bestSolution.Solution) > 0 {
		tmpSolution := [][]toolkit.Step{}
		for _, ambulancePath := range bestSolution.Solution {
			i := 0
			pathStr := ""
			for _, step := range ambulancePath {
				i++
				if pathStr == "" {
					pathStr = fmt.Sprintf("%d H%d", step.StartTime, -step.From)
				}

				if step.To < 0 {
					// returned to hospital
					pathStr += fmt.Sprintf(" H%d", -step.To)
					break
				} else {
					pathStr += fmt.Sprintf(" P%d", step.To+1)
				}
			}

			if i < len(ambulancePath) {
				tmpSolution = append(tmpSolution, ambulancePath[i:])
			}
			f.WriteString(pathStr + "\n")
		}
		bestSolution.Solution = tmpSolution

	}
	f.Sync()

}
