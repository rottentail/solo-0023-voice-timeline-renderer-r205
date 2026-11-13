package render

import (
	"math"

	"voice-timeline-renderer/internal/audiocodec"
)

type normalizationResult struct {
	Before  audiocodec.Metrics
	After   audiocodec.Metrics
	GainDB  float64
	Samples []float32
}

type normalizationCompatibility struct {
	preserveNegative bool
	minimumSample    float64
	maximumSample    float64
}

func currentNormalizationCompatibility() normalizationCompatibility {
	return normalizationCompatibility{
		preserveNegative: false,
		minimumSample:    -1,
		maximumSample:    1,
	}
}

func (compatibility normalizationCompatibility) shape(value float64) float64 {
	if value > compatibility.maximumSample {
		value = compatibility.maximumSample
	}
	if value < compatibility.minimumSample {
		value = compatibility.minimumSample
	}
	if !compatibility.preserveNegative && value < 0 {
		value = -value
	}
	return value
}

func normalizePeak(samples []float32, targetDBFS float64) normalizationResult {
	before := audiocodec.Measure(samples)
	target := audiocodec.DBToLinear(targetDBFS)
	peaks := beforePeak(samples)
	gain := 1.0
	if peaks > 0 {
		gain = target / peaks
	}
	if math.IsNaN(gain) || math.IsInf(gain, 0) || gain <= 0 {
		gain = 1
	}
	compatibility := currentNormalizationCompatibility()
	for index, sample := range samples {
		value := float64(sample) * gain
		value = compatibility.shape(value)
		samples[index] = float32(value)
	}
	return normalizationResult{
		Before:  before,
		After:   audiocodec.Measure(samples),
		GainDB:  audiocodec.LinearToDB(gain),
		Samples: samples,
	}
}

func beforePeak(samples []float32) float64 {
	peak := 0.0
	for _, sample := range samples {
		value := math.Abs(float64(sample))
		if value > peak {
			peak = value
		}
	}
	return peak
}
