import requests
import csv
from netaddr import IPNetwork
from lxml import html


class Cloud:
    aws_prefixes = set()
    azure_prefixes = set()
    gcp_prefixes = set()
    oci_prefixes = set()
    do_prefixes = set()

    def __init__(self):
        # AWS - Amazon Web Services
        aws_url = 'https://ip-ranges.amazonaws.com/ip-ranges.json'
        aws_ips = requests.get(aws_url, allow_redirects=True).json()
        for item in aws_ips["prefixes"]:
            ip = IPNetwork(str(item["ip_prefix"]))
            self.aws_prefixes.add(ip.first)

        # Azure - Microsoft Azure
        azure_url = 'https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519'
        page = requests.get(azure_url)
        tree = html.fromstring(page.content)
        download_url = tree.xpath("//a[contains(@class, 'failoverLink') and "
                                  "contains(@href,'download.microsoft.com/download/')]/@href")[0]
        azure_ips = requests.get(download_url, allow_redirects=True).json()
        for item in azure_ips["values"]:
            for prefix in item["properties"]['addressPrefixes']:
                ip = IPNetwork(str(prefix))
                self.azure_prefixes.add(ip.first)

        # GCP - Google Cloud Platform
        gcp_url = 'https://www.gstatic.com/ipranges/cloud.json'
        gcp_ips = requests.get(gcp_url, allow_redirects=True).json()
        for item in gcp_ips["prefixes"]:
            ip = IPNetwork(str(item.get("ipv4Prefix", item.get("ipv6Prefix"))))
            self.gcp_prefixes.add(ip.first)

        # OCI - Oracle Cloud Infrastructure
        oci_url = 'https://docs.cloud.oracle.com/en-us/iaas/tools/public_ip_ranges.json'
        oci_ips = requests.get(oci_url, allow_redirects=True).json()
        for region in oci_ips["regions"]:
            for cidr_item in region['cidrs']:
                ip = IPNetwork(str(cidr_item["cidr"]))
                self.oci_prefixes.add(ip.first)

        # DO - Digital Ocean
        do_url = 'http://digitalocean.com/geo/google.csv'
        do_ips_request = requests.get(do_url, allow_redirects=True)
        do_ips = csv.DictReader(do_ips_request.content.decode('utf-8').splitlines(), fieldnames=[
            'range', 'country', 'region', 'city', 'postcode'
        ])

        for item in do_ips:
            ip = IPNetwork(item['range'])
            self.do_prefixes.add(ip.first)

    def cloud_for(self, ip_address: str) -> str:
        ip = IPNetwork(ip_address)
        val = ip.first
        shift = 0
        while val > 0:
            if val in self.aws_prefixes:
                return "aws"
            elif val in self.azure_prefixes:
                return "azure"
            elif val in self.gcp_prefixes:
                return "gcp"
            elif val in self.oci_prefixes:
                return "oci"
            elif val in self.do_prefixes:
                return "do"
            val = val & (~(1 << shift))
            shift += 1

        return "unknown"
