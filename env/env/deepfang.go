package env

import (
	"golib/neuralnet"
	"golib/rule"
	"math"
	"math/rand"
)

type DeepFangPolicy interface {
	Select(logits []float32) rule.Direction
}

type RandomDeepFangPolicy struct {
}

func (l RandomDeepFangPolicy) Select(logits []float32) rule.Direction {
	if len(logits) != 4 {
		panic("logits must have exactly 4 elements for directions")
	}
	probabilities := softmax(logits)
	actionIndex := weightedChoice(probabilities)
	return rule.Direction(actionIndex)
}

type GreedyDeepFangPolicy struct {
}

func (l GreedyDeepFangPolicy) Select(logits []float32) rule.Direction {
	if len(logits) != 4 {
		panic("logits must have exactly 4 elements for directions")
	}
	maxIndex := 0
	maxValue := logits[0]
	for i, p := range logits {
		if p > maxValue {
			maxValue = p
			maxIndex = i
		}
	}
	return rule.Direction(maxIndex)
}

type DeepFang struct {
	model         neuralnet.Model
	policy        DeepFangPolicy
	onDataCollect func(data *Experience)
}

func NewDeepFang(model neuralnet.Model, policy DeepFangPolicy, onDataCollect func(experience *Experience)) *DeepFang {
	return &DeepFang{model, policy, onDataCollect}
}

func (d *DeepFang) OnStartCallback(ctx *GameContext) {
}

func (d *DeepFang) OnMoveCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake) rule.Direction {
	input := neuralnet.GetInputTensor(*board)
	actor, critic, err := d.model.Predict(input)
	if err != nil {
		panic("model prediction failed: " + err.Error())
	}
	action := rule.Up
	if actor != nil {
		action = d.policy.Select(actor)
	}
	ctx.SetValue("action", action)
	ctx.SetValue("input", board)
	ctx.SetValue("actor", actor)
	ctx.SetValue("critic", critic)
	return action
}

func (d *DeepFang) OnTurnEndCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake, result uint32) {
	input, _ := ctx.GetValue("input")
	actor, _ := ctx.GetValue("actor")
	critic, _ := ctx.GetValue("critic")
	action, _ := ctx.GetValue("action")
	stepData := &Experience{
		Board:  input.(*rule.Board),
		Actor:  actor.([]float32),
		Critic: critic.(float32),
		Action: int(action.(rule.Direction)),
		Reward: calculateReward(board, snake, result),
		Done:   false,
	}
	d.onDataCollect(stepData)
}

func weightedChoice(probs []float32) int {
	var total float32
	for _, p := range probs {
		total += p
	}

	r := rand.Float32() * total

	var accum float32
	for i, p := range probs {
		accum += p
		if r < accum {
			return i
		}
	}
	return len(probs) - 1
}

func softmax(logits []float32) []float32 {
	maxLogit := logits[0]
	for _, v := range logits {
		if v > maxLogit {
			maxLogit = v
		}
	}

	// exp の計算 & 合計
	exps := make([]float64, len(logits))
	var sum float64
	for i, v := range logits {
		exps[i] = math.Exp(float64(v - maxLogit))
		sum += exps[i]
	}

	// 正規化して確率に変換
	probs := make([]float32, len(logits))
	for i := range logits {
		probs[i] = float32(exps[i] / sum)
	}
	return probs
}

func calculateReward(board *rule.Board, snake *rule.Snake, result uint32) float32 {
	reward := 0.0
	if (result & rule.TurnResultMaskAlive) == rule.TurnResultAlive {
		reward += 0.1 // Small positive reward for surviving the turn
	}
	if (result & rule.TurnResultMaskFood) == rule.TurnResultFoodEaten {
		reward += 1.0 // Positive reward for eating food
	}
	if (result & rule.TurnResultMaskResult) == rule.TurnResultDead {
		reward += -5.0 // Negative reward for dying
	}
	return float32(reward)
}
