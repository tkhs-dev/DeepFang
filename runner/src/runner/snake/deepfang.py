from typing import Any

import torch

from runner.snake.snake import ISnake, Direction
from lib.model.src.model.model import EvaluatorModel

class DeepFang(ISnake):
    def __init__(self, model: EvaluatorModel):
        self.myid = None
        self.snakes = None
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.model = model.to(self.device)
        self.width, self.height = model.size
        self.num_snakes = model.num_snakes
        super().__init__()

    def state_to_tensors(self, state: dict) -> tuple[torch.Tensor, torch.Tensor]:
        board_tensor = torch.zeros((self.num_snakes * 3 + 2, self.height + 2, self.width + 2), dtype=torch.float32)
        features_tensor = torch.zeros((2 * self.num_snakes,), dtype=torch.float32)
        you = state['you']
        # my head
        board_tensor[0, you['body'][0]['y'] + 1, you['body'][0]['x'] + 1] = 1.0
        for body in you['body']:
            board_tensor[1, body['y'] + 1, body['x'] + 1] = 1.0
        # my tail
        tail = you['body'][-1]
        board_tensor[2, tail['y'] + 1, tail['x'] + 1] = 1.0

        # features
        features_tensor[0] = you['health'] / 100.0
        features_tensor[1] = len(you['body']) / (self.width * self.height)

        for i, snake in enumerate(state['board']['snakes']):
            if snake['id'] == you['id']:
                continue
            # head
            board_tensor[i * 3 + 3, snake['body'][0]['y'] + 1, snake['body'][0]['x'] + 1] = 1.0
            # body
            for body in snake['body']:
                board_tensor[i * 3 + 4, body['y'] + 1, body['x'] + 1] = 1.0
            # tail
            tail = snake['body'][-1]
            board_tensor[i * 3 + 5, tail['y'] + 1, tail['x'] + 1] = 1.0
            # features
            features_tensor[2 + i * 2] = snake['health'] / 100.0
            features_tensor[2 + i * 2 + 1] = len(snake['body']) / (self.width * self.height)
        # food
        for food in state['board']['food']:
            board_tensor[self.num_snakes * 3, food['y'] + 1, food['x'] + 1] = 1.0
        # walls
        board_tensor[self.num_snakes * 3 + 1, 0, :] = 1.0
        board_tensor[self.num_snakes * 3 + 1, -1, :] = 1.0
        board_tensor[self.num_snakes * 3 + 1, :, 0] = 1.0
        board_tensor[self.num_snakes * 3 + 1, :, -1] = 1.0

        return board_tensor.unsqueeze(0), features_tensor.unsqueeze(0)

    def info(self) -> dict[str, Any]:
        return {
            "apiversion": "1",
            "author": "DeepFang",
            "color": "#ff0000",
            "head": "default",
            "tail": "default",
        }

    def on_start(self, state):
        self.myid = state['you']['id']
        self.snakes = list(filter(lambda x: x['id'] != self.myid, [s for s in state['board']['snakes']]))
        num_snakes = len(self.snakes) + 1
        assert num_snakes == self.num_snakes, f"Model was trained for {self.num_snakes} snakes, but got {num_snakes} snakes."
        assert state['board']['width'] == self.width and state['board']['height'] == self.height, f"Model was trained for board size ({self.width}, {self.height}), but got ({state['board']['width']}, {state['board']['height']})."

    def on_move(self, state) -> Direction:
        board_tensor, features_tensor = self.state_to_tensors(state)
        with torch.no_grad():
            policy_logits, _ = self.model(board_tensor.to(self.device), features_tensor.to(self.device))
            policy = torch.softmax(policy_logits, dim=-1)
            # パーセント表示(小数第二位) パーセントもつける
            print("Policy:", ", ".join([f"{p*100:.2f}%" for p in policy.squeeze().cpu().numpy()]))
            action = torch.argmax(policy, dim=-1).item()
        return Direction.from_index(action)

    def on_end(self, state):
        pass