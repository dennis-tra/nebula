import re
from typing import Optional

known_agents = [
    "go-ipfs",
    "kubo",
    "hydra-booster",
    "storm",
    "ioi",
    "iroh"
]


def agent_name(agent_version) -> str:
    for agent in known_agents:
        if agent in agent_version:
            if agent == "go-ipfs":
                return "kubo"
            return agent
    return "other"


def kubo_version(agent_version) -> Optional[str]:
    match = re.match(r"(go-ipfs|kubo)\/(\d+\.+\d+\.\d+)(.*)?\/", agent_version)
    if match is None:
        return None

    return match.group(2)
