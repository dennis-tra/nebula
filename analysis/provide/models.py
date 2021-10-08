from dataclasses import dataclass
from datetime import datetime
import json
from os import path
from typing import List, Any, TypeVar, Callable
import dateutil.parser

T = TypeVar("T")


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


def from_float(x: Any) -> float:
    assert isinstance(x, (float, int)) and not isinstance(x, bool)
    return float(x)


def from_datetime(x: Any) -> datetime:
    return dateutil.parser.parse(x)


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


def to_float(x: Any) -> float:
    assert isinstance(x, float)
    return x


@dataclass
class MeasurementInfo:
    started_at: datetime
    ended_at: datetime
    content_id: str
    provider_id: str
    provider_dist: str
    requester_id: str
    requester_dist: str
    peer_order: List[str]

    @staticmethod
    def from_dict(obj: Any) -> 'MeasurementInfo':
        assert isinstance(obj, dict)
        return MeasurementInfo(
            from_datetime(obj.get("StartedAt")),
            from_datetime(obj.get("EndedAt")),
            from_str(obj.get("ContentID")),
            from_str(obj.get("ProviderID")),
            from_str(obj.get("ProviderDist")),
            from_str(obj.get("RequesterID")),
            from_str(obj.get("RequesterDist")),
            from_list(from_str, obj.get("PeerOrder"))
        )

    def to_dict(self) -> dict:
        return {
            "StartedAt": self.started_at.isoformat(),
            "EndedAt": self.ended_at.isoformat(),
            "ContentID": from_str(self.content_id),
            "ProviderID": from_str(self.provider_id),
            "ProviderDist": from_str(self.provider_dist),
            "RequesterID": from_str(self.requester_id),
            "RequesterDist": from_str(self.requester_dist),
            "PeerOrder": from_list(from_str, self.peer_order),
        }


@dataclass
class Span:
    rel_start: float
    duration_s: float
    start: datetime
    end: datetime
    peer_id: str
    type: str
    error: str

    @staticmethod
    def from_dict(obj: Any) -> 'Span':
        assert isinstance(obj, dict)
        return Span(
            from_float(obj.get("RelStart")),
            from_float(obj.get("DurationS")),
            from_datetime(obj.get("Start")),
            from_datetime(obj.get("End")),
            from_str(obj.get("PeerID")),
            from_str(obj.get("Type")),
            from_str(obj.get("Error")),
        )

    def to_dict(self) -> dict:
        return {
            "RelStart": to_float(self.rel_start),
            "DurationS": to_float(self.duration_s),
            "Start": self.start.isoformat(),
            "End": self.end.isoformat(),
            "PeerID": from_str(self.peer_id),
            "Type": from_str(self.type),
            "Error": from_str(self.error)
        }


@dataclass
class PeerInfo:
    id: str
    agent_version: str
    xor_distance: str
    rel_discovered_at: float
    discovered_at: datetime
    discovered_from: str

    @staticmethod
    def from_dict(obj: Any) -> 'PeerInfo':
        assert isinstance(obj, dict)
        return PeerInfo(
            from_str(obj.get("ID")),
            from_str(obj.get("AgentVersion")),
            from_str(obj.get("XORDistance")),
            from_float(obj.get("RelDiscoveredAt")),
            from_datetime(obj.get("DiscoveredAt")),
            from_str(obj.get("DiscoveredFrom")),
        )

    def to_dict(self) -> dict:
        return {
            "ID": from_str(self.id),
            "AgentVersion": from_str(self.agent_version),
            "XORDistance": from_str(self.xor_distance),
            "RelDiscoveredAt": to_float(self.rel_discovered_at),
            "DiscoveredAt": self.discovered_at.isoformat(),
            "DiscoveredFrom": from_str(self.discovered_from)
        }


@dataclass
class Measurement:
    info: MeasurementInfo
    peer_infos: dict[str, PeerInfo]
    provider_spans: dict[str, list[Span]]
    requester_spans: dict[str, list[Span]]

    @staticmethod
    def from_location(loc: str, prefix: str):

        # Load measurement information
        measurement_info: MeasurementInfo
        with open(path.join(loc, f"{prefix}_measurement_info.json")) as f:
            measurement_info = MeasurementInfo.from_dict(json.load(f))

        # Load peer information
        peer_infos: dict[str, PeerInfo] = {}
        with open(path.join(loc, f"{prefix}_peer_infos.json")) as f:
            data = json.load(f)
            for key in data:
                peer_infos[key] = PeerInfo.from_dict(data[key])

        # Load provider spans
        provider_spans: dict[str, list[Span]] = {}
        with open(path.join(loc, f"{prefix}_provider_spans.json")) as f:
            data = json.load(f)
            for key in data:
                provider_spans[key] = []
                for span_dict in data[key]:
                    provider_spans[key] += [Span.from_dict(span_dict)]

        # Load requester spans
        requester_spans: dict[str, list[Span]] = {}
        with open(path.join(loc, f"{prefix}_requester_spans.json")) as f:
            data = json.load(f)
            for key in data:
                requester_spans[key] = []
                for span_dict in data[key]:
                    requester_spans[key] += [Span.from_dict(span_dict)]

        return Measurement(measurement_info, peer_infos, provider_spans, requester_spans)
