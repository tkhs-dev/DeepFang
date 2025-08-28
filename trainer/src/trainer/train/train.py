import os
import warnings
from typing import List, Tuple

from lib.model.src.model.model import EvaluatorModel

import torch.optim as optim
import torch.nn.functional as F
import torch

from trainer.client.client import EnvArgs, Experience


class PPOAgent:
    """PPOエージェント"""
    def __init__(self, lr: float = 3e-4,
                 gamma: float = 0.99, eps_clip: float = 0.2,
                 value_coef: float = 0.5, entropy_coef: float = 0.01,
                 max_grad_norm: float = 0.5):

        self.gamma = gamma
        self.eps_clip = eps_clip
        self.value_coef = value_coef
        self.entropy_coef = entropy_coef
        self.max_grad_norm = max_grad_norm

        # ネットワークの初期化
        self.policy = EvaluatorModel((11, 11), 1)
        self.optimizer = optim.Adam(self.policy.parameters(), lr=lr)

        # デバイスの設定
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        # self.device = torch.device("cpu")
        self.policy.to(self.device)

    def get_model_bytes(self) -> bytes:
        self.policy.to_onnx("temp.onnx")
        with open("temp.onnx", "rb") as f:
            model_bytes = f.read()
        return model_bytes

    def compute_gae(self, experiences: List[Experience], lam: float = 0.95) -> Tuple[torch.Tensor, torch.Tensor]:
        """Generalized Advantage Estimation (GAE)を計算"""
        rewards = torch.stack([exp.rewards for exp in experiences])
        dones   = torch.stack([exp.dones for exp in experiences])
        values  = torch.stack([exp.values for exp in experiences])

        T = len(experiences)
        advantages = torch.zeros(T, dtype=torch.float32)
        returns    = torch.zeros(T, dtype=torch.float32)

        gae = 0.0
        next_value = 0.0

        for t in reversed(range(T)):
            if dones[t]:  # エピソード終端なら bootstrap はしない
                next_value = 0.0
                gae = 0.0
                next_nonterminal = 0.0
            else:
                next_nonterminal = 1.0

            delta = rewards[t] + self.gamma * next_value * next_nonterminal - values[t]
            gae = delta + self.gamma * lam * next_nonterminal * gae
            advantages[t] = gae
            returns[t] = advantages[t] + values[t]

            # 次のステップのために更新
            next_value = values[t]

        return advantages, returns
    flg = False
    def update(self, experiences: List[Experience], n_epochs: int = 4, batch_size: int = 64):
        """PPO更新"""
        # データを結合
        board_states = torch.cat([exp.states[0] for exp in experiences])
        feature_states = torch.cat([exp.states[1] for exp in experiences])
        actions = torch.cat([exp.actions for exp in experiences])
        old_log_probs = torch.cat([exp.log_probs for exp in experiences])

        # GAEを計算
        advantages, returns = self.compute_gae(experiences)

        # アドバンテージの正規化
        advantages = (advantages - advantages.mean()) / (advantages.std() + 1e-8)

        # デバイスに移動
        board_states = board_states.to(self.device)
        feature_states = feature_states.to(self.device)
        actions = actions.to(self.device)
        old_log_probs = old_log_probs.to(self.device)
        advantages = advantages.to(self.device)
        returns = returns.to(self.device)

        # 複数エポック更新
        for _ in range(n_epochs):
            # バッチサイズでシャッフル
            indices = torch.randperm(len(feature_states))

            for start_idx in range(0, len(feature_states), batch_size):
                end_idx = start_idx + batch_size
                batch_indices = indices[start_idx:end_idx]

                batch_board_states = board_states[batch_indices]
                batch_feature_states = feature_states[batch_indices]
                batch_actions = actions[batch_indices]
                batch_old_log_probs = old_log_probs[batch_indices]
                batch_advantages = advantages[batch_indices]
                batch_returns = returns[batch_indices]

                # 現在の方策で評価
                new_log_probs, values, entropy = self.policy.evaluate_actions(
                    batch_board_states, batch_feature_states, batch_actions
                )

                # 重要度比を計算
                ratio = torch.exp(new_log_probs - batch_old_log_probs)

                # PPOクリップ損失
                surr1 = ratio * batch_advantages
                surr2 = torch.clamp(ratio, 1 - self.eps_clip, 1 + self.eps_clip) * batch_advantages
                actor_loss = -torch.min(surr1, surr2).mean()

                # 価値関数損失
                with warnings.catch_warnings(record=True) as w:
                    warnings.simplefilter("always")  # 全ての警告を捕捉
                    critic_loss = F.mse_loss(values.squeeze(-1), batch_returns)

                    if len(w) > 0 or self.flg:
                        print(f"Warning during MSE loss computation")
                        print(start_idx, end_idx)
                        print(values.squeeze(-1), values.squeeze())
                        print(w[0].message)
                        self.flg = True


                # エントロピー損失
                entropy_loss = -entropy.mean()

                # 総損失
                total_loss = actor_loss + self.value_coef * critic_loss + self.entropy_coef * entropy_loss

                # 更新
                self.optimizer.zero_grad()
                total_loss.backward()
                torch.nn.utils.clip_grad_norm_(self.policy.parameters(), self.max_grad_norm)
                self.optimizer.step()

class PPOTrainer:
    """PPO訓練クラス"""
    def __init__(self, env_client, agent: PPOAgent, env_args: EnvArgs):
        self.env_client = env_client
        self.agent = agent
        self.env_args = env_args

    def train(self, n_iterations: int, steps_per_iteration: int = 1024):
        if not os.path.exists("./models"):
            os.makedirs("./models")
        """訓練ループ"""
        for iteration in range(n_iterations):
            print(f"Iteration {iteration + 1}/{n_iterations}")

            # モデルをONNX形式で取得
            model_bytes = self.agent.get_model_bytes()

            # 経験を収集
            experiences = self.env_client.collect_experience(
                env_args=self.env_args,
                models=[model_bytes],
                worker_num=1,
                steps=steps_per_iteration,
                chunk_size=10000
            )

            # PPO更新
            self.agent.update(experiences=experiences, n_epochs=4, batch_size=64)

            # 統計情報の計算
            total_reward = sum(exp.rewards.sum().item() for exp in experiences)
            avg_reward = total_reward / len(experiences) if experiences else 0

            print(f"Average Reward: {avg_reward:.2f}, Total Reward: {total_reward:.2f}")

            # 定期的にモデルを保存
            if (iteration + 1) % 50 == 0:
                torch.save(self.agent.policy.state_dict(), f"./models/ppo_model_iter{iteration + 1}.pth")
                print(f"Saved model at iteration {iteration + 1}")

        print("Training completed!")