import time
from multiprocessing.managers import State

import torch

import trainer.client.grpc.env_pb2
from lib.model.src.model.model import EvaluatorModel
from trainer.client.client import EnvClient, EnvArgs
from trainer.client.converter import from_grpc_state
from trainer.train.train import PPOAgent, PPOTrainer

class FakePoint:
    def __init__(self, x, y):
        self.x = x
        self.y = y

class FakeSnake:
    body = [FakePoint(1,1), FakePoint(1,2), FakePoint(1,3)]

class FakeBoardState:
    food = [FakePoint(5, 5)]
    snakes = [FakeSnake()]

class FakeFeatureState:
    features = [0.5, 0.5]

class FakeState:
    board_state = FakeBoardState()
    feature_state = FakeFeatureState()


def cli():
    torch.set_printoptions(edgeitems=1000, precision=9)
    # model = EvaluatorModel((11,11), 1)
    # model.to_onnx()
    # # load onnx model data as bytes
    # with open("output.onnx", "rb") as f:
    #     model_data = f.read()
    # print(f"Model data size: {len(model_data)} bytes")
    # client = EnvClient()
    # env_args = EnvArgs(
    #     width=11,
    #     height=11,
    #     food_spawn_chance=15,
    #     minimum_food=1,
    #     snakes=["model:0"],
    # )
    # start_time = time.time()
    # collected = client.collect_experience(env_args, [model_data], worker_num=1, steps=1, chunk_size=500)
    # client.close()
    # print([action for exp in collected for action in exp.actions.tolist()])
    # end_time = time.time()
    #
    # # Print elapsed time in milliseconds
    # elapsed_time = (end_time - start_time) * 1000
    # print(f"Elapsed time: {elapsed_time:.2f} ms")

    # model = EvaluatorModel((6,6), 1)
    # board_state, feature_state = from_grpc_state(
    #     FakeState(),
    #     6,
    #     6,
    #     1,
    # )
    # print("board_state")
    # print(board_state)
    # print("feature_state")
    # print(feature_state)
    # result = model.forward(board_state, feature_state)
    # model.to_onnx("output.onnx")
    # print(result)

    env_args = EnvArgs(
        width=11,
        height=11,
        food_spawn_chance=25,
        minimum_food=3,
        snakes=["model:0"],
    )
    client = EnvClient()

    # エージェントの作成
    agent = PPOAgent()

    # EnvClientは外部から渡される想定
    trainer = PPOTrainer(client, agent, env_args)
    trainer.train(n_iterations=5000)