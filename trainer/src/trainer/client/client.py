from dataclasses import dataclass
from typing import List, Tuple

import grpc
import torch
from torch.distributions import Categorical
import torch.nn.functional as F

import trainer
from trainer.client.converter import from_grpc_board, from_grpc_state
from trainer.client.grpc.env_pb2 import CollectExperienceRequest
from trainer.client.grpc.env_pb2_grpc import EnvServiceStub

@dataclass
class EnvArgs:
    width: int
    height: int
    food_spawn_chance: int
    minimum_food: int
    snakes: list[str]

@dataclass
class Experience:
    """PPOに必要な経験データ"""
    states: Tuple[torch.Tensor, torch.Tensor]
    actions: torch.Tensor
    rewards: torch.Tensor
    dones: torch.Tensor
    log_probs: torch.Tensor
    values: torch.Tensor

class EnvClient:
    def __init__(self, host='localhost', port=8080):
        self.channel = grpc.insecure_channel(f'{host}:{port}')
        self.stub = EnvServiceStub(self.channel)

    def close(self):
        self.channel.close()

    def collect_experience(self, env_args: EnvArgs, models: list[bytes], worker_num: int, steps: int, chunk_size: int = 1000) -> List[Experience]:
        args = trainer.client.grpc.env_pb2.EnvArgs(
            board_width=env_args.width,
            board_height=env_args.height,
            food_spawn_chance=env_args.food_spawn_chance,
            minimum_food=env_args.minimum_food,
            snakes=env_args.snakes,
        )
        request = CollectExperienceRequest(args=args, worker_num=worker_num, models=models, step_num=steps, chunk_size=chunk_size)
        result = []
        for response in self.stub.CollectExperience(request):
            experiences: List[trainer.client.grpc.env_pb2.Experience] = response.experiences
            for exp in experiences:
                states = from_grpc_state(exp.state, env_args.width, env_args.height, len(env_args.snakes))
                actions = torch.tensor([exp.action], dtype=torch.int64)
                rewards = torch.tensor([exp.reward], dtype=torch.float32)
                dones = torch.tensor([exp.done], dtype=torch.float32)
                log_probs = torch.tensor([exp.log_prob], dtype=torch.float32)
                values = torch.tensor([exp.critic], dtype=torch.float32)
                result.append(Experience(
                    states=states,
                    actions=actions,
                    rewards=rewards,
                    dones=dones,
                    log_probs=log_probs,
                    values=values,
                ))
        return result

