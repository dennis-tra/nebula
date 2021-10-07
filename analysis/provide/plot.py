from models import *
import json

measurement_info: MeasurementInfo
peer_infos: dict[str, PeerInfo] = {}

with open('2021-10-07T18:38:49_measurement_info.json') as f:
    measurement_info = MeasurementInfo.from_dict(json.load(f))

with open('2021-10-07T18:38:49_peer_infos.json') as f:
    data = json.load(f)
    for key in data.keys():
        peer_infos[key] = PeerInfo.from_dict(data[key])

print(peer_infos)
