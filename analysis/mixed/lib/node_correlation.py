import numpy as np
import pytz
import matplotlib.pyplot as plt


def get_target_by_daytime(start, end, window, shift, duration, tz):
    start = start.astimezone(tz)
    end = end.astimezone(tz)
    startNom = start.replace(hour=0, minute=0, second=0, microsecond=0)
    remove_bar = round((start - startNom).total_seconds() / window) + 1
    num_bar = round((end - startNom).total_seconds() / window)
    resOn = np.zeros(shape=(num_bar))
    resOff = np.ones(shape=(num_bar))
    resDefault = np.zeros(shape=(num_bar))
    on_bar = round(shift / window)
    off_bar = round((shift + duration) / window)
    while on_bar < num_bar:
        resOn[on_bar:off_bar] = 1
        resOff[on_bar:off_bar] = 0
        on_bar += round(86400 / window)
        off_bar += round(86400 / window)
    return resOn[remove_bar:], resOff[remove_bar:], resDefault[remove_bar:]


def get_up_time_correlation(conn, start, end, window, shift, duration, peer_ids, tz):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    targetOn, targetOff, targetDefault = get_target_by_daytime(start, end, window, shift, duration, tz)
    # Initialise peer session map
    peer_session_map = dict()
    for peer_id in peer_ids:
        peer_session_map[peer_id] = np.copy(targetDefault)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT peer_id, first_successful_dial, last_successful_dial
        FROM sessions
        WHERE created_at < %s AND updated_at > %s AND peer_id IN (%s)
        """ % ("%s", "%s", ','.join(['%s'] * len(peer_ids))),
        (end, start) + tuple(peer_ids)
    )
    for peer_id, on, off in cur.fetchall():
        on_bar = round((on - start).total_seconds() / window)
        off_bar = round((off - start).total_seconds() / window)
        peer_session_map[peer_id][on_bar:off_bar] = 1
        pass
    # Calculate correlation
    resOn = dict()
    resOff = dict()
    for id, target in peer_session_map.items():
        resOn[id] = np.correlate(target, targetOn)[0]
        resOff[id] = np.correlate(target, targetOff)[0]
    return resOn, resOff
