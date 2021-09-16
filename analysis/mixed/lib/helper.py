# Helper function, copied from nebula crawler analysis.
def node_address(maddr):
    try:
        return maddr.value_for_protocol(0x04)
    except:
        pass
    return maddr.value_for_protocol(0x29)


# Helper function, copied from nebula crawler analysis.
def parse_maddr_str(maddr_str):
    """
    The following line parses a row like:
      {/ip6/::/tcp/37374,/ip4/151.252.13.181/tcp/37374}
    into
      ['/ip6/::/tcp/37374', '/ip4/151.252.13.181/tcp/37374']
    """
    return maddr_str.replace("{", "").replace("}", "").split(",")
