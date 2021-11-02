import matplotlib.pyplot as plt

def get_millis(duration):
    if duration.endswith("µs"):
        return float(duration[:-2]) / 1000
    elif duration.endswith("ms"):
        return float(duration[:-2])
    elif duration.endswith("s"):
        if "m" in duration:
            return int(duration[0]) * 60 * 1000 + get_millis(duration[2:])
        return float(duration[:-2]) * 1000
    else:
        return -1

mode = 0
pvd = []
ret = []
for line in list(open("./storm.time")):
    line = line.strip("\n")
    if line == "publish":
        mode = 0
    if line == "fetch":
        mode = 1
    else:
        ms = get_millis(line)
        if  ms > 0:
            if mode == 0:
                pvd.append(ms)
            else:
                ret.append(ms)
# Generate graphs
plt.rc('font', size=8)
plt.rcParams["figure.figsize"] = (10,6)
# Overall pvd latency
plt.clf()
plt.hist(list(pvd), bins=20, density=False)
plt.title("Storm nodes put provider record response time")
plt.xlabel("Response time (ms)")
plt.ylabel("Count")
plt.savefig("./figs/storm_put_response_time.png")
# Overall pvd latency trim
trim = list(pvd)
trim = [x for x in trim if x <= 1]
plt.clf()
plt.hist(trim, bins=100, density=False)
plt.title("Storm nodes put provider record response time (Trimmed)")
plt.xlabel("Response time (ms)")
plt.ylabel("Count")
plt.savefig("./figs/storm_put_response_time_trim.png")
# Overall ret latency
plt.clf()
plt.hist(list(ret), bins=20, density=False)
plt.title("Storm nodes fetch provider record response time")
plt.xlabel("Response time (ms)")
plt.ylabel("Count")
plt.savefig("./figs/storm_fetch_response_time.png")
# Overall ret latency trim
trim = list(ret)
trim = [x for x in trim if x <= 1500]
plt.clf()
plt.hist(trim, bins=100, density=False)
plt.title("Storm nodes fetch provider record response time (Trimmed)")
plt.xlabel("Response time (ms)")
plt.ylabel("Count")
plt.savefig("./figs/storm_fetch_response_time_trim.png")