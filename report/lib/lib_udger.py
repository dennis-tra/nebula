import sqlite3
import numpy as np
from dataclasses import dataclass
from typing import Optional


@dataclass
class Datacenter:
    id: int
    name: str
    name_code: str
    homepage: str


class UdgerClient:

    def __init__(self):
        self.conn = sqlite3.connect("udgerdb_v3.dat")

    def get_datacenter(self, id: int) -> Optional[Datacenter]:
        if id is None or np.isnan(id):
            return None

        cur = self.conn.cursor()
        cur.execute(f"SELECT * FROM udger_datacenter_list WHERE id = {id}")
        result = cur.fetchall()
        if result is None or len(result) != 1:
            return None

        return Datacenter(result[0][0], result[0][1], result[0][2], result[0][3])

    def close(self):
        self.conn.close()
