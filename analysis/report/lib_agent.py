from typing import Optional

known_agents = [
    "go-ipfs",
    "hydra-booster",
    "storm",
    "ioi"
]

go_ipfs_version_mappings = [
    ["go-ipfs/0.4", "0.4.x"],
    ["go-ipfs/0.7", "0.7.x"],
    ["go-ipfs/0.8", "0.8.x"],
    ["go-ipfs/0.9", "0.9.x"],
    ["go-ipfs/0.10", "0.10.x"],
    ["go-ipfs/0.11", "0.11.x"],
    ["go-ipfs", "others"],
]


def agent_name(agent_version) -> str:
    """
    client_name returns the name of the process that is
    interacting with the DHT.
    """
    for agent in known_agents:
        if agent in agent_version:
            return agent
    return "others"


def go_ipfs_version(agent_version) -> Optional[str]:
    """
    Helper function to get the go IPFS agent minor version
    """
    for mapping in go_ipfs_version_mappings:
        if mapping[0] in agent_version:
            return mapping[1]
    return None


def go_ipfs_v08_version(agent_version):
    """
    Helper function to trim agent version of go-ipfs 0.8.x
    """
    if "go-ipfs/0.8" not in agent_version:
        return None
    elif "go-ipfs/0.8.0/" in agent_version:
        return agent_version[8:]
    elif "go-ipfs/0.8.0/16615d7" in agent_version:
        return agent_version[8:]
    elif "go-ipfs/0.8.0/ce693d7" in agent_version:
        return agent_version[8:]
    elif "go-ipfs/0.8.0/48f94e2" in agent_version:
        return agent_version[8:]
    else:
        return "others"
