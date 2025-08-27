from typing import Tuple

import torch
from torch import nn
from torch.distributions import Categorical
import torch.nn.functional as F


class EvaluatorModel(nn.Module):
    def __init__(self, boardsize: Tuple[int, int], num_snakes: int):
        super(EvaluatorModel, self).__init__()

        feature_input = 2 * num_snakes

        self.cnn_input_shape = (3 * num_snakes + 1, boardsize[0], boardsize[1])
        self.feature_input_shape = (feature_input,)
        self.actor_output_shape = (4,)
        self.critic_output_shape = (1,)

        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        self.cnn = nn.Sequential(
            nn.Conv2d(3*num_snakes + 1, 16, kernel_size=3, stride=1, padding=1),
            nn.ReLU(),
            nn.Conv2d(16, 32, kernel_size=3, stride=1, padding=1),
            nn.ReLU(),
            nn.Conv2d(32, 64, kernel_size=3, stride=1, padding=1),
            nn.ReLU(),
        )
        self.mlp_features = nn.Sequential(
            nn.Linear(feature_input, feature_input),
            nn.ReLU(),
        )
        self.fc_shared = nn.Sequential(
            nn.Linear(64 * boardsize[0] * boardsize[1] + feature_input, 128),
            nn.ReLU(),
            nn.Linear(128, 128),
        )
        self.fc_actor = nn.Linear(128, 4)

        self.fc_critic = nn.Linear(128, 1)

    def forward(self, board, features):
        """
        board: 盤面入力 [B, C, H, W]
        features: 数値特徴入力 [B, num_features]
        戻り値:
            policy_logits: [B, 4] 各行動のlogit
            value: [B, 1] 状態価値
        """
        board_feat = self.cnn(board)
        feat_vec = self.mlp_features(features)
        board_feat = board_feat.view(board_feat.size(0), -1)
        x = torch.cat((board_feat, feat_vec), dim=1)
        x = self.fc_shared(x)

        policy_logits = self.fc_actor(x)
        value = self.fc_critic(x)

        return policy_logits, value

    def get_action_and_value(self, board_states, feature_states):
        """行動と価値を取得"""
        action_logits, value = self.forward(board_states, feature_states)
        dist = Categorical(logits=F.softmax(action_logits, dim=-1))
        action = dist.sample()
        log_prob = dist.log_prob(action)
        return action, log_prob, value

    def evaluate_actions(self, board_states, feature_states, actions):
        """行動を評価（PPO更新用）"""
        action_logits, values = self.forward(board_states, feature_states)
        dist = Categorical(logits=F.softmax(action_logits, dim=-1))
        log_probs = dist.log_prob(actions)
        entropy = dist.entropy()
        return log_probs, values, entropy

    def to_onnx(self, filename="output.onnx", opset_version=17):
        import torch.onnx
        dummy_board = torch.randn(1, *self.cnn_input_shape).to(self.device)
        dummy_features = torch.randn(1, *self.feature_input_shape).to(self.device)

        torch.onnx.export(
            self,
            (dummy_board, dummy_features),
            filename,
            input_names=["input_cnn", "input_features"],
            output_names=["output_actor", "output_critic"],
            dynamic_axes={
                "input_cnn": {0: "batch_size"},
                "input_features": {0: "batch_size"},
                "output_actor": {0: "batch_size"},
                "output_critic": {0: "batch_size"},
            },
            opset_version=opset_version,
        )
