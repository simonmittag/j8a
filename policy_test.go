package j8a

import (
	"testing"
)

func TestEmptyPolicyResolves(t *testing.T) {
	p := Policy{}
	want := "default"
	got := p.resolveLabel()
	if got != want {
		t.Errorf("resolved incorrect policy label, got %v, want %v", got, want)
	}
}

func TestPolicyResolveLabelWithWeightOne(t *testing.T) {
	p := Policy{
		LabelWeight{
			Label:  "green",
			Weight: 1,
		},
		LabelWeight{
			Label:  "blue",
			Weight: 0,
		},
	}

	//we don't want this to fail because of chance, so we repeat within reason.
	for i := 0; i < 10; i++ {
		want := "green"
		got := p.resolveLabel()
		if got != want {
			t.Errorf("resolved incorrect policy label, got %v, want %v", got, want)
		}
	}
}

func TestPolicyResolveLabelWithDistributedWeight(t *testing.T) {
	p := Policy{
		LabelWeight{
			Label:  "green",
			Weight: 0.8,
		},
		LabelWeight{
			Label:  "blue",
			Weight: 0.2,
		},
	}

	gotGreen := 0
	gotBlue := 0

	//we don't want this to fail because of chance, so we repeat within reason.
	for i := 0; i < 10000; i++ {
		got := p.resolveLabel()
		if got == "green" {
			gotGreen++
		}
		if got == "blue" {
			gotBlue++
		}
	}

	wantGreen := 0.8
	wantBlue := 0.2
	pcgotblue := float64(gotBlue) / float64(gotBlue+gotGreen)
	pcgotgreen := float64(gotGreen) / float64(gotBlue+gotGreen)

	//we allow 2% deviation in a test sample of 10k before failing the test.
	if pcgotblue < 0.18 || pcgotblue > 0.22 {
		t.Errorf("incorrect amount of blue labels resolved, wanted percentage %v, got %v", wantBlue, pcgotblue)
	}
	if pcgotgreen < 0.78 || pcgotgreen > 0.82 {
		t.Errorf("incorrect amount of greens labels resolved, wanted percentage %v, got %v", wantGreen, pcgotgreen)
	}
}
