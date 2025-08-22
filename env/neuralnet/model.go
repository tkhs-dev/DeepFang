package neuralnet

import (
	"fmt"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

type InputData struct {
	CNNInput     []float32 `json:"cnn_input"`     // Input data for CNN
	FeatureInput []float32 `json:"feature_input"` // General input data
}

type Model interface {
	// Predict takes an input and returns the predicted output.
	Predict(input *InputData) ([]float32, float32, error)
	Close() error // Close releases any resources held by the model.
}

type OnnxModel struct {
	session            *ort.AdvancedSession
	cnnInputTensor     *ort.Tensor[float32] // CNN input tensor
	featureInputTensor *ort.Tensor[float32] // General input tensor
	actorOutputTensor  *ort.Tensor[float32] // Actor output tensor
	criticOutputTensor *ort.Tensor[float32] // Critic output tensor
	width              int
	height             int
	numSnakes          int        // Number of snakes in the game
	mutex              sync.Mutex // Mutex for thread safety
}

func (o *OnnxModel) Predict(input *InputData) ([]float32, float32, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.cnnInputTensor.ZeroContents()
	copy(o.cnnInputTensor.GetData(), input.CNNInput)
	o.featureInputTensor.ZeroContents()
	copy(o.featureInputTensor.GetData(), input.FeatureInput)
	err := o.session.Run()
	if err != nil {
		return nil, 0, err
	}
	actorOutput := o.actorOutputTensor.GetData()
	criticOutput := o.criticOutputTensor.GetData()
	result := make([]float32, len(actorOutput))
	copy(result, actorOutput)
	return result, criticOutput[0], nil
}

func (o *OnnxModel) Close() error {
	o.cnnInputTensor.Destroy()
	o.featureInputTensor.Destroy()
	o.actorOutputTensor.Destroy()
	o.criticOutputTensor.Destroy()
	if o.session != nil {
		err := o.session.Destroy()
		if err != nil {
			return fmt.Errorf("failed to close session: %v", err)
		}
		o.session = nil
	}
	return nil
}

func NewOnnxModel(modelData []byte, numSnakes int, width int, height int) (*OnnxModel, error) {
	cnnInputShape := ort.NewShape(1, int64(numSnakes*3+1), int64(height), int64(width))
	featureInputShape := ort.NewShape(1, int64(numSnakes*2))

	cnnInputTensor, err := ort.NewEmptyTensor[float32](cnnInputShape)
	if err != nil {
		return nil, fmt.Errorf("NewEmptyTensor: %v", err)
	}
	featureInputTensor, err := ort.NewEmptyTensor[float32](featureInputShape)
	if err != nil {
		cnnInputTensor.Destroy()
		return nil, fmt.Errorf("NewEmptyTensor: %v", err)
	}

	actorOutputShape := ort.NewShape(1, int64(4)) // 4 possible actions
	criticOutputShape := ort.NewShape(1, int64(1))
	actorOutputTensor, err := ort.NewEmptyTensor[float32](actorOutputShape)
	if err != nil {
		cnnInputTensor.Destroy()
		featureInputTensor.Destroy()
		return nil, fmt.Errorf("NewEmptyTensor: %v", err)
	}
	criticOutputTensor, err := ort.NewEmptyTensor[float32](criticOutputShape)
	if err != nil {
		cnnInputTensor.Destroy()
		featureInputTensor.Destroy()
		actorOutputTensor.Destroy()
		return nil, fmt.Errorf("NewEmptyTensor: %v", err)
	}
	session, err := ort.NewAdvancedSessionWithONNXData(
		modelData,
		[]string{"input_cnn", "input_features"},
		[]string{"output_actor", "output_critic"},
		[]ort.Value{cnnInputTensor, featureInputTensor},
		[]ort.Value{actorOutputTensor, criticOutputTensor},
		nil,
	)
	if err != nil {
		cnnInputTensor.Destroy()
		featureInputTensor.Destroy()
		actorOutputTensor.Destroy()
		criticOutputTensor.Destroy()
		return nil, err
	}
	return &OnnxModel{
		session:            session,
		cnnInputTensor:     cnnInputTensor,
		featureInputTensor: featureInputTensor,
		actorOutputTensor:  actorOutputTensor,
		criticOutputTensor: criticOutputTensor,
		width:              width,
		height:             height,
		numSnakes:          numSnakes}, nil
}
