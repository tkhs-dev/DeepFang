from typing import Tuple

import torch

from trainer.client.grpc.env_pb2 import BoardState
from trainer.client.grpc.env_pb2 import State


def from_grpc_state(state: State, width: int, height: int, num_snakes: int) -> Tuple[torch.Tensor, torch.Tensor]:
    """gRPCのStateをPyTorchのTensorに変換する"""
    board = from_grpc_board(state.board_state, width, height, num_snakes).unsqueeze(0)
    # additional features
    features = torch.tensor(state.feature_state.features, dtype=torch.float32).unsqueeze(0)
    return board, features

def from_grpc_board(board: BoardState, width: int, height: int, num_snakes: int) -> torch.Tensor:
    """gRPCのBoardをPyTorchのTensorに変換する"""
    # empty board
    state = torch.zeros((num_snakes * 3 + 2, height + 2, width + 2), dtype=torch.float32)
    # snakes
    for i, snake in enumerate(board.snakes):
        if not snake.body:
            continue
        # head
        state[i * 3, snake.body[0].y + 1, snake.body[0].x + 1] = 1.0
        # body
        for segment in snake.body:
            state[i * 3 + 1, segment.y + 1, segment.x + 1] = 1.0
        # tail
        tail = snake.body[-1]
        state[i * 3 + 2, tail.y + 1, tail.x + 1] = 1.0
    # food
    for food in board.food:
        state[num_snakes * 3, food.y + 1, food.x + 1] = 1.0
    # walls (borders)
    state[-1, 0, :] = 1.0
    state[-1, -1, :] = 1.0
    state[-1, :, 0] = 1.0
    state[-1, :, -1] = 1.0

    return state

def print_board(state: torch.Tensor):
    """状態を可視化する"""
    c = ['.', 'F'] + [f'H' for i in range((state.shape[0] - 1) // 3)] + [f'B' for i in range((state.shape[0] - 1) // 3)] + [f'T' for i in range((state.shape[0] - 1) // 3)]
    h, w = state.shape[1], state.shape[2]
    board = [['.' for _ in range(w)] for _ in range(h)]
    for i in range(state.shape[0]):
        for y in range(h):
            for x in range(w):
                if state[i, y, x] > 0:
                    board[y][x] = c[i]
    for row in board:
        print(' '.join(row))
    print()