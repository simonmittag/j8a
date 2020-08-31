package j8a

import (
	"math/rand"
	"sort"
)

//LabelWeight describes routing to labels
type LabelWeight struct {
	Label  string
	Weight float64
}

//Policy defines an array of LabelWeights used for routing
type Policy []LabelWeight

func (policy Policy) Len() int {
	return len(policy)
}

func (policy Policy) Swap(i, j int) {
	policy[i], policy[j] = policy[j], policy[i]
}
func (policy Policy) Less(i, j int) bool {
	return policy[i].Weight < policy[j].Weight
}

//resolve a label inside a policy
func (policy Policy) resolveLabel() string {
	dice := rand.Float64()
	sort.Sort(Policy(policy))
	var cw []float64
	//add up cumulative weights 0 < w < 1
	for i := 1; i <= len(policy); i++ {
		var sum float64 = 0
		for i := range policy[0:i] {
			sum += policy[i].Weight
		}
		cw = append(cw, sum)
	}
	//select a random number 0 < dice < 1 inside distribution
	//and map label
	for i := 0; i < len(policy); i++ {
		if dice < cw[i] {
			return policy[i].Label
		}
	}
	return "default"
}
