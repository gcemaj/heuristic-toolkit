package toolkit

import (
	"math"
	"math/rand"

	"github.com/muesli/clusters"
)

type Ant interface {
	ComputeFitness() float64
	ComputeSolution(alpha, beta float64, graph *Graph) [][]Step
	Reset()
}

type Location struct {
	Street int
	Avenue int
}

type Patient struct {
	Loc        Location
	TimeToLive int
}

type Ambulance struct {
	CurrentLocation  int
	StartingLocation int
	CurrentTime      int
	Patients         []int
}

type HospitalAnt struct {
	Patients   []Patient
	Hospitals  []Location
	Visited    []bool
	Ambulances []Ambulance
	Saved      []int
	MaxTime    int
}

func (this Location) Distance(point clusters.Coordinates) float64 {
	coords := this.Coordinates()
	return math.Abs(coords[0]-point[0]) + math.Abs(coords[1]-point[1])
}

func (this Location) Coordinates() clusters.Coordinates {
	return clusters.Coordinates{
		float64(this.Street),
		float64(this.Avenue),
	}
}

func (ha *HospitalAnt) Reset() {
	ha.Saved = make([]int, len(ha.Patients))
	ha.Visited = make([]bool, len(ha.Patients))
	for _, a := range ha.Ambulances {
		a.CurrentLocation = a.StartingLocation
		a.CurrentTime = 0
		a.Patients = []int{}
	}
}

func (ha *HospitalAnt) ComputeFitness() float64 {
	score := 0.0
	for id, time := range ha.Saved {
		if time > 0 && time <= ha.Patients[id].TimeToLive {
			score += 1.0
		}
	}

	return score
}

func ComputeDistance(a, b Location) int {
	return int(math.Abs(float64(a.Avenue)-float64(b.Avenue))) + int(math.Abs(float64(a.Street)-float64(b.Street)))
}

func (ha *HospitalAnt) getNextLocation(ambulance Ambulance, alpha, beta float64, graph *Graph) (int, int) {
	// Compute closets Hospital
	minHospitalDist := math.MaxInt
	closestHospital := 0

	if ambulance.CurrentLocation >= 0 {
		for hi := range ha.Hospitals {
			dist := graph.Edges[-hi-1][ambulance.CurrentLocation].Distance
			if dist < minHospitalDist {
				minHospitalDist = dist
				closestHospital = hi
			}
		}
	}

	// Get next patient that will die
	minToLive := math.MaxInt
	for _, p := range ha.Patients {
		if p.TimeToLive < minToLive {
			minToLive = p.TimeToLive
		}
	}

	// If we have 4 patients, go to the nearest hospital
	if len(ambulance.Patients) == 4 {
		return -closestHospital - 1, minHospitalDist
	}

	scores := map[int]float64{}
	maxScore := 0.0

	for nextId, edge := range graph.Edges[ambulance.CurrentLocation] {
		// We already visited that patient
		if ha.Visited[nextId] {
			continue
		}
		// Compute closets Hospital from next patient
		minNextHospitalDist := math.MaxInt
		for hi := range ha.Hospitals {
			dist := graph.Edges[-hi-1][nextId].Distance
			if dist < minNextHospitalDist {
				minNextHospitalDist = dist
			}
		}

		if ambulance.CurrentTime+edge.Distance+minHospitalDist+2 > minToLive {
			continue
		}

		timeLeft := ha.Patients[nextId].TimeToLive - ambulance.CurrentTime
		if timeLeft < 0 {
			continue
		}

		exploitation := math.Pow(edge.Phermone, alpha)
		exploration := math.Pow((1.0/(float64(timeLeft*timeLeft)))+(1.0/(float64(edge.Distance))), beta)
		score := exploration * exploitation

		scores[nextId] = score
		maxScore += score
	}

	toss := rand.Float64()
	cumulative := 0.0
	for id, score := range scores {
		weight := (score / maxScore)
		if toss <= (weight + cumulative) {
			return id, graph.Edges[ambulance.CurrentLocation][id].Distance
		}
		cumulative += weight
	}
	if minHospitalDist == math.MaxInt {
		return -closestHospital - 1, 0
	}
	return -closestHospital - 1, minHospitalDist
}

func (ha *HospitalAnt) ComputeSolution(alpha, beta float64, graph *Graph) [][]Step {
	solution := [][]Step{}
	for _, ambulance := range ha.Ambulances {
		ambulancePath := []Step{}
		for ambulance.CurrentTime < ha.MaxTime {
			nextId, dist := ha.getNextLocation(ambulance, alpha, beta, graph)

			if dist == 0 {
				break
			}

			startTime := ambulance.CurrentTime
			startLocation := ambulance.CurrentLocation

			ambulance.CurrentTime += dist
			ambulance.CurrentLocation = nextId

			if nextId >= 0 {
				// Pick up another patient
				ambulance.CurrentTime += 1
				ha.Visited[nextId] = true
				ambulance.Patients = append(ambulance.Patients, nextId)

			} else {
				// Go to hospital
				ambulance.CurrentTime += 1

				for _, patientId := range ambulance.Patients {
					ha.Saved[patientId] = ambulance.CurrentTime
				}

				ambulance.Patients = []int{}
			}

			ambulancePath = append(ambulancePath, Step{
				From:      startLocation,
				To:        nextId,
				StartTime: startTime,
				EndTime:   ambulance.CurrentTime,
			})
		}
		solution = append(solution, ambulancePath)
	}

	return solution
}
