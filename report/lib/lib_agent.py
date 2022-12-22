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
    if agent_version == "go-ipfs/0.8.0/48f94e2":
        return "storm*"

    for agent in known_agents:
        if agent_version.startswith(agent):
            if agent == "go-ipfs":
                return "kubo"
            return agent
    return "other"


def kubo_version(agent_version) -> Optional[str]:
    if agent_version == "go-ipfs/0.8.0/48f94e2":
        return None

    match = re.match(r"(go-ipfs|kubo)\/(\d+\.+\d+\.\d+)(.*)?\/", agent_version)
    if match is None:
        return None

    return match.group(2)
