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
	ctx.SetValue("feature", input.FeatureInput)
	ctx.SetValue("board", board)
	ctx.SetValue("log_prob", logProb(actor, int(action)))
	ctx.SetValue("critic", critic)
	return action
}

func (d *DeepFang) OnTurnEndCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake, turn int, result uint32) {
	alreadyDone, found := ctx.GetValue("done")
	if found && alreadyDone.(bool) {
		return
	}
	action, _ := ctx.GetValue("action")
	feature, _ := ctx.GetValue("feature")
	b, _ := ctx.GetValue("board")
	log_prob, _ := ctx.GetValue("log_prob")
	critic, _ := ctx.GetValue("critic")
	done := (result & rule.TurnResultMaskResult) != rule.TurnResultContinue
	if done {
		ctx.SetValue("done", true)
		println(turn, calculateReward(board, snake, result), result&rule.TurnResultMaskCause == rule.TurnResultStarvation)
	}
	stepData := &Experience{
		State: &State{
			Board:   b.(*rule.Board),
			Feature: feature.([]float32),
		},
		LogProb: log_prob.(float32),
		Critic:  critic.(float32),
		Action:  int32(action.(rule.Direction)),
		Reward:  calculateReward(board, snake, result),
		Done:    done,
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

func logProb(logits []float32, action int) float32 {
	probs := softmax(logits)
	if action < 0 || action >= len(probs) {
		panic("action index out of range")
	}
	return float32(math.Log(float64(probs[action]) + 1e-10)) // Add small value to avoid log(0)
}

func calculateReward(board *rule.Board, snake *rule.Snake, result uint32) float32 {
	reward := 0.0
	if (result & rule.TurnResultMaskAlive) == rule.TurnResultAlive {
		reward += 0.1 // Small positive reward for surviving the turn
	} else {
		if snake.Length > 2 && snake.Head.Equal(snake.Body[2]) {
			// 自殺ペナルティ
			reward += -3.0
		} else {
			const (
				C     = 3.0 // 最大のペナルティ
				alpha = 0.5 // 軽減係数 (0=軽減なし, 1=最大軽減)
			)

			boardSize := float64(board.Width * board.Height)
			p := float64(snake.Length) / boardSize

			reward += -C * (1.0 - alpha*p)
		}
	}
	if (result & rule.TurnResultMaskFood) == rule.TurnResultFoodEaten {
		const (
			maxHealth = 100
			threshold = 30
			a         = 1   // 体力ゼロ時の最大正報酬
			b         = 0.3 // 体力満タン時の最大負報酬（小さい値にする）
		)

		h := float64(snake.Health)

		if h < threshold {
			// 体力が少ないとき → 正報酬 (0 ~ +a)
			reward += a * (1.0 - h/threshold)
		}
		// 体力が多いとき → 負報酬 (0 ~ -b)
		reward += -b * (h - threshold) / (maxHealth - threshold)
	}
	return float32(reward)
}
