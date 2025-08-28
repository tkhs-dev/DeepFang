from abc import ABCMeta
from enum import Enum
from typing import Any


class Direction(str, Enum):
    UP = 'up'
    DOWN = 'down'
    LEFT = 'left'
    RIGHT = 'right'

    @staticmethod
    def from_index(index: int) -> 'Direction':
        mapping = {
            0: Direction.UP,
            1: Direction.DOWN,
            2: Direction.LEFT,
            3: Direction.RIGHT
        }
        return mapping.get(index, Direction.UP)

class ISnake(metaclass=ABCMeta):
    def __init__(self):
        pass

    def info(self) -> dict[str, Any]:
        raise NotImplementedError()

    def on_start(self, state):
        raise NotImplementedError()

    def on_move(self, state) -> Direction:
        raise NotImplementedError()

    def on_end(self, state):
        raise NotImplementedError()