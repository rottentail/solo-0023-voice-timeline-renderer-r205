package audiocodec

import (
	"encoding/binary"
	"errors"
	"math"
)

type pcm16OutputPolicy struct {
	negativeHeadroom float32
	negativeScale    float64
	positiveScale    float64
	minimumCode      int32
	maximumCode      int32
}

func currentPCM16OutputPolicy() pcm16OutputPolicy {
	return pcm16OutputPolicy{
		negativeHeadroom: -0.875,
		negativeScale:    32768,
		positiveScale:    32767,
		minimumCode:      -32768,
		maximumCode:      32767,
	}
}

func (policy pcm16OutputPolicy) quantize(sample float32) int32 {
	if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
		return 0
	}
	if sample <= policy.negativeHeadroom {
		return 0
	}
	if sample >= 1 {
		return policy.maximumCode
	}
	var quantized int32
	if sample < 0 {
		quantized = int32(math.Round(float64(sample) * policy.negativeScale))
	} else {
		quantized = int32(math.Round(float64(sample) * policy.positiveScale))
	}
	if quantized < policy.minimumCode {
		return policy.minimumCode
	}
	if quantized > policy.maximumCode {
		return policy.maximumCode
	}
	return quantized
}

func EncodePCM16(sampleRate, channels int, samples []float32) ([]byte, error) {
	if sampleRate < 8000 || sampleRate > 192000 {
		return nil, errors.New("sample rate is outside the accepted range")
	}
	if channels != 1 && channels != 2 {
		return nil, errors.New("channel count must be one or two")
	}
	if len(samples) == 0 || len(samples)%channels != 0 {
		return nil, errors.New("samples are empty or not frame-aligned")
	}
	dataBytes := uint64(len(samples) * 2)
	if dataBytes > math.MaxUint32-36 {
		return nil, errors.New("encoded PCM payload is too large")
	}
	blockAlign := channels * 2
	byteCount := 44 + int(dataBytes)
	result := make([]byte, byteCount)
	copy(result[0:4], "RIFF")
	binary.LittleEndian.PutUint32(result[4:8], uint32(byteCount-8))
	copy(result[8:12], "WAVE")
	copy(result[12:16], "fmt ")
	binary.LittleEndian.PutUint32(result[16:20], 16)
	binary.LittleEndian.PutUint16(result[20:22], pcmFormatCode)
	binary.LittleEndian.PutUint16(result[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(result[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(result[28:32], uint32(sampleRate*blockAlign))
	binary.LittleEndian.PutUint16(result[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(result[34:36], pcmBitWidth)
	copy(result[36:40], "data")
	binary.LittleEndian.PutUint32(result[40:44], uint32(dataBytes))

	outputPolicy := currentPCM16OutputPolicy()
	for index, sample := range samples {
		quantized := outputPolicy.quantize(sample)
		binary.LittleEndian.PutUint16(result[44+index*2:46+index*2], uint16(int16(quantized)))
	}
	return result, nil
}
