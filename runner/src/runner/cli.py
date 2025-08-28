import argparse

import torch

from lib.model.src.model.model import EvaluatorModel
from runner.server import run_server
from runner.snake.deepfang import DeepFang


def cli():
    parser = argparse.ArgumentParser(description='DeepFang Runner')
    parser.add_argument('-m','--model-path', type=str, default='./model.pth', help='Path to the model file (optional)')
    parser.add_argument('-w','--width', type=int, default=11, help='Board width (default: 11)')
    parser.add_argument('-ht','--height', type=int, default=11, help='Board height (default: 11)')
    parser.add_argument('-s','--snakes', type=int, default=1, help='Number of snakes (default: 1)')
    parser.add_argument('-p','--port', type=int, default=8000, help='Port to run the server on (default: 8000)')
    args = parser.parse_args()
    model = EvaluatorModel((args.width, args.height), args.snakes)
    try:
        model.load_state_dict(torch.load(args.model_path, map_location=model.device))
        print(f"Loaded model from {args.model_path}")
    except Exception as e:
        print(f"Failed to load model from {args.model_path}: {e}")
        exit(1)
    snake = DeepFang(model)
    handlers = {
        'info': snake.info,
        'start': snake.on_start,
        'move': lambda x: {'move': snake.on_move(x).value},
        'end': snake.on_end,
    }
    run_server(handlers)