import matplotlib.pyplot as plt

loc_map = {"DE": [737,407],"CZ": [734,433],"NL": [721,402],"GB": [693,398],"BE": [720,412],"CN": [1183,487],"FR": [709,431],"US": [281,468],"PL": [779,401],
    "TR": [847,469],"IE": [666,399],"CH": [733,432],"MY": [1120,623],"SE": [764,338],"AT": [758,427],"RU": [880,376],"HK": [1173,543],"SG": [1129,635],"BR": [485,678],
    "ES": [682,462],"KR": [1228,483],"TW": [1198,542],"JP": [1274,483],"PA": [362,608],"CL": [410,749],"CA": [208,373],"FI": [799,341],"BS": [378,537],"LT": [798,386],
    "AR": [429,789],"UA": [827,414],"VN": [1145,590],"AU": [1314,779],"VE": [428,607],"MX": [281,546],"LV": [805,376],"EE": [805,364],"GR": [790,469],"MT": [762,485],
    "IN": [1037,550],"CY": [836,492],"BG": [802,452],"DK": [736,380],"PT": [665,467],"AM": [665,467],"BY": [816,397],"HU": [778,428],"MD": [818,431],"RO": [799,437],
    "AX": [784,363],"PK": [975,521],"ID": [1198,649],"SI": [789,459],"IR": [920,500],"EC": [376,644],"AE": [924,541],"IT": [747,448],"AO": [768,692],"BA": [771,445],
    "GE": [876,455],"GU": [876,455],"NO": [734,349],"NG": [729,602],"UY": [465,780],"SN": [638,576],"HN": [342,578],"MO": [1166,549],"TH": [1115,574],"LU": [727,419],
    "RE": [912,731],"CO": [396,624],"SK": [778,421],"AL": [782,461],"JM": [380,567],"PE": [383,679],"IL": [843,507],"HT": [399,564],"ZA": [795,771],"MK": [789,458],
    "SR": [468,622],"NZ": [1422,809],"HR": [762,445],"RS": [783,444],"BH": [911,534],"DO": [410,562],"SA": [879,537],"SC": [926,642],"MA": [672,502],"KW": [895,515],
    "KZ": [985,419],"CR": [352,599],"PY": [456,737],"GH": [693,609],"BD": [1070,539],"MM": [1092,548],"LB": [844,500],"SV": [333,585],"UZ": [962,458],"PR": [427,567],
    "BN": [1174,623],"PH": [1214,592],"IS": [620,328],"MU": [924,727]}


def plot_line(pos1, pos2, lat, count, max_lat, max_count):
    width = 1
    color = 'b'
    ratio = count * 1.0 / max_count
    if ratio <= 0.1:
        width = 0.01
    if ratio > 0.1 and ratio <= 0.25:
        width = 0.05
    elif ratio > 0.25 and ratio <= 0.5:
        width = 0.1
    elif ratio > 0.5 and ratio <= 0.75:
        width = 0.15
    else:
        width = 0.3
    ratio = lat / max_lat
    if ratio <= 0.1:
        color = 'darkgreen'
    elif ratio > 0.1 and ratio <= 0.2:
        color = 'green'
    elif ratio > 0.2 and ratio <= 0.3:
        color = 'yellow'
    elif ratio > 0.3 and ratio <= 0.4:
        color = 'orange'
    elif ratio > 0.4 and ratio <= 0.5:
        color = 'sandybrown'
    elif ratio > 0.5 and ratio <= 0.6:
        color = 'orangered'
    elif ratio > 0.6 and ratio <= 0.7:
        color = 'red'
    elif ratio > 0.7 and ratio <= 0.8:
        color = 'darkred'
    elif ratio > 0.8 and ratio <= 0.9:
        color = 'maroon'
    else:
        color = 'maroon'
    plt.plot([pos1[0], pos2[0]], [pos1[1], pos2[1]], linewidth=width, color=color)


plt.rcParams["figure.figsize"] = (20, 13)
img = plt.imread("empty_map.jpeg")
plt.imshow(img, cmap='Greys_r')
plt.title("Latency map")

latency1 = dict()
min_count = 0
max_count = 0
min_lat = 0
max_lat = 0
with open("./region-latencies/london.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency1[loc] = (lat, count)

latency2 = dict()
with open("./region-latencies/mumbai.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency2[loc] = (lat, count)

latency3 = dict()
with open("./region-latencies/ohio.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency3[loc] = (lat, count)

latency4 = dict()
with open("./region-latencies/oregon.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency4[loc] = (lat, count)

latency5 = dict()
with open("./region-latencies/paris.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency5[loc] = (lat, count)

latency6 = dict()
with open("./region-latencies/sa.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency6[loc] = (lat, count)

latency7 = dict()
with open("./region-latencies/seoul.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency7[loc] = (lat, count)

latency8 = dict()
with open("./region-latencies/sydney.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency8[loc] = (lat, count)

latency9 = dict()
with open("./region-latencies/tokyo.latency") as f:
    line = f.readline()
    line = line.split('{')[1].split('}')[0]
    for entry in line.split("),"):
        loc = entry.split(":")[0].strip().strip('\'')
        count = int(entry.split(",")[1][1:].strip(")"))
        lat = float(entry.split("(")[1].split(',')[0])
        if max_count == 0:
            max_count = min_count = count
            max_lat = min_lat = lat
        else:
            if count > max_count:
                max_count = count
            if count < min_count:
                min_count = count
            if lat > max_lat:
                max_lat = lat
            if lat < min_lat:
                min_lat = lat
        latency9[loc] = (lat, count)

london = [696, 406]
for loc, res in latency1.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(london, loc_map[loc], lat, count, max_lat, max_count)

mumbai = [1003, 561]
for loc, res in latency2.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(mumbai, loc_map[loc], lat, count, max_lat, max_count)

ohio = [360, 464]
for loc, res in latency3.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(ohio, loc_map[loc], lat, count, max_lat, max_count)

oregon = [196, 444]
for loc, res in latency4.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(oregon, loc_map[loc], lat, count, max_lat, max_count)

paris = [709, 426]
for loc, res in latency5.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(paris, loc_map[loc], lat, count, max_lat, max_count)

sa = [531, 707]
for loc, res in latency6.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(sa, loc_map[loc], lat, count, max_lat, max_count)

seoul = [1225, 479]
for loc, res in latency7.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(seoul, loc_map[loc], lat, count, max_lat, max_count)

sydney = [1313, 785]
for loc, res in latency8.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(sydney, loc_map[loc], lat, count, max_lat, max_count)

tokyo = [1270, 487]
for loc, res in latency9.items():
    if loc == "unknown":
        continue
    if loc not in loc_map:
        continue
    lat = res[0]
    count = res[1]
    plot_line(tokyo, loc_map[loc], lat, count, max_lat, max_count)

plt.savefig("./region-latencies/latency-map.png")