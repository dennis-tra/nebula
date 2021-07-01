import numpy as np

max_interval_s = 60*60*24
min_interval_s = 30


def new_interval(current):
    new_interval_s = current * 1.2
    if new_interval_s > max_interval_s:
        return max_interval_s
    elif new_interval_s < min_interval_s:
        return min_interval_s
    else:
        return new_interval_s

interval = 0
for i in range(40):
    interval = new_interval(interval)
    if interval > 60*60:
        print(f"{interval//3600} h")
    elif interval > 60:
        print(f"{interval/60} min")
    else:
        print(f"{interval} s")

# 30 s
# 45.0 s
# 1.12 min
# 1.68 min
# 2.53 min
# 3.79 min
# 5.69 min
# 8.54 min
# 12.81 min
# 19.22 min
# 28.83 min
# 43.24 min
# 1.08 h
# 1.62 h
# 2.43 h
# 3.64 h
# 5.47 h
# 8.21 h
# 12.31 h
# 18.47 h
# 24.0 h
# 24.0 h
