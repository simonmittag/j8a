package server

import (
	"math/rand"
	"sort"
)

//LabelWeights type for sorting
type LabelWeights []LabelWeight

//LabelWeight describes routing to labels
type LabelWeight struct {
	Label  string
	Weight float64
}

func (labelWeights LabelWeights) Len() int {
	return len(labelWeights)
}

func (labelWeights LabelWeights) Swap(i, j int) {
	labelWeights[i], labelWeights[j] = labelWeights[j], labelWeights[i]
}
func (labelWeights LabelWeights) Less(i, j int) bool {
	return labelWeights[i].Weight < labelWeights[j].Weight
}

//Policy defines an array of LabelWeights used for routing
type Policy struct {
	LabelWeights LabelWeights
}

func (policy Policy) resolveLabel() string {
	dice := rand.Float64()
	sort.Sort(LabelWeights(policy.LabelWeights))
	var l = len(policy.LabelWeights)
	var cw []float64
	for i := 1; i <= l; i++ {
		var sum float64 = 0
		for i := range policy.LabelWeights[0:i] {
			sum += policy.LabelWeights[i].Weight
		}
		cw = append(cw, sum)
	}
	for i := 0; i < l; i++ {
		if dice < cw[i] {
			return policy.LabelWeights[i].Label
		}
	}
	return "none"
}
