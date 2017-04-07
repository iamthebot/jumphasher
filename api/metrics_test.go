package main

import (
	"encoding/json"
	"math/rand"
	"testing"
)

func TestMetricsEngine_AddDuration(t *testing.T) {
	var me MetricsEngine

	//add 10 million requests. We need a large number to test the asymptotic error characteristics of our calculation
	delays := make([]float64, 10000000) //holds delays in seconds. Beware, this will use up 80MAB ram
	for i := 0; i < len(delays); i++ {
		delays[i] = rand.Float64() //generates in [0.0,1.0) interval
		me.AddDuration(int64(delays[i] * 1000000000))
	}

	// calculate sum of delays so we can get sample mean
	// we use Kahan compensated summation otherwise floating point errors add up FAST!
	var sum float64 = 0.0  //running sum of delays
	var comp float64 = 0.0 //floating point error compensator
	for i := 0; i < len(delays); i++ {
		shifted := delays[i] - comp
		part_sum := sum + shifted
		comp = (part_sum - sum) - shifted //update running compensation term
		sum = part_sum
	}
	mean := uint64(sum / float64(len(delays)) * 1000) //convert from seconds to milliseconds
	snap := me.MSSnapshot()                           //snapshot of metrics

	if uint64(len(delays)) != snap.Total {
		t.Errorf("Expected count: %d Actual Count: %d", len(delays), snap.Total)
	}
	if mean != snap.Average {
		t.Errorf("Expected Mean: %d Actual Mean: %d", mean, snap.Average)
	}
}

func TestMSMetrics_MarshalJson(t *testing.T) {
	var me MetricsEngine

	//add 10 million requests
	for i := 0; i < 10000000; i++ {
		me.AddDuration(int64(rand.Float64() * 1000000000))
	}
	snap := me.MSSnapshot()
	j, err := snap.MarshalJson()
	if err != nil {
		t.Error(err)
	}

	//now deserialize it
	var snapDsrz MSMetrics
	err = json.Unmarshal(j, &snapDsrz)
	if err != nil {
		t.Error(err)
	}
	if snapDsrz.Total != snap.Total {
		t.Errorf("Expected: %d Actual: %d", snap.Total, snapDsrz.Total)
	} else if snapDsrz.Total != snap.Total {
		t.Errorf("Expected: %d Actual: %d", snap.Average, snapDsrz.Average)
	}
}
