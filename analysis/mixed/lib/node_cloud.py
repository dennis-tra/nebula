import requests
from netaddr import IPNetwork
from multiaddr import Multiaddr
from lxml import html
from threading import Thread
import csv
import helper


# get_cloud gets the cloud info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the cloud info of these peer ids.
def get_cloud(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT id, multi_addresses
        FROM peers
        WHERE id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    # aws
    aws_url = 'https://ip-ranges.amazonaws.com/ip-ranges.json'
    aws_ips = requests.get(aws_url, allow_redirects=True).json()
    aws_prefixes = set()
    for item in aws_ips["prefixes"]:
        ip = IPNetwork(str(item["ip_prefix"]))
        aws_prefixes.add(ip.first)
    # azure
    azure_url = 'https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519'
    page = requests.get(azure_url)
    tree = html.fromstring(page.content)
    download_url = tree.xpath("//a[contains(@class, 'failoverLink') and "
                                "contains(@href,'download.microsoft.com/download/')]/@href")[0]
    azure_ips = requests.get(download_url, allow_redirects=True).json()
    azure_prefixes = set()
    for item in azure_ips["values"]:
        for prefix in item["properties"]['addressPrefixes']:
            ip = IPNetwork(str(prefix))
            azure_prefixes.add(ip.first)
    # gcp
    gcp_url = 'https://www.gstatic.com/ipranges/cloud.json'
    gcp_ips = requests.get(gcp_url, allow_redirects=True).json()
    gcp_prefixes = set()
    for item in gcp_ips["prefixes"]:
        ip = IPNetwork(str(item.get("ipv4Prefix", item.get("ipv6Prefix"))))
        gcp_prefixes.add(ip.first)
    # oci
    oci_url = 'https://docs.cloud.oracle.com/en-us/iaas/tools/public_ip_ranges.json'
    oci_ips = requests.get(oci_url, allow_redirects=True).json()
    oci_prefixes = set()
    for region in oci_ips["regions"]:
        for cidr_item in region['cidrs']:
            ip = IPNetwork(str(cidr_item["cidr"]))
            oci_prefixes.add(ip.first)
    # do
    do_url = 'http://digitalocean.com/geo/google.csv'
    do_ips_request = requests.get(do_url, allow_redirects=True)
    do_ips = csv.DictReader(do_ips_request.content.decode('utf-8').splitlines(), fieldnames=[
        'range', 'country', 'region', 'city', 'postcode'
    ])
    do_prefixes = set()
    for item in do_ips:
        ip = IPNetwork(item['range'])
        do_prefixes.add(ip.first)

    for id, maddr_strs in cur.fetchall():
        found = False
        for maddr_str in maddr_strs:
            maddr = Multiaddr(maddr_str)
            try:
                address = helper.node_address(maddr)
                ip = IPNetwork(address)
                val = ip.first
                shift = 0
                while val > 0:
                    if val in aws_prefixes:
                        res[id] = "aws"
                        found = True
                        break
                    elif val in azure_prefixes:
                        res[id] = "azure"
                        found = True
                        break
                    elif val in gcp_prefixes:
                        res[id] = "gcp"
                        found = True
                        break
                    elif val in oci_prefixes:
                        res[id] = "oci"
                        found = True
                        break
                    elif val in do_prefixes:
                        res[id] = "do"
                        found = True
                        break
                    val = val & (~(1 << shift))
                    shift += 1
            except:
                pass
            if found:
                break
        if not found:
            res[id] = "unknown"
    return res
