import time
import requests
from .utils import wait_for_new_blocks

def test_authz(cronos):
    updated = """
[program:cronos_777-1-node0]
autostart = true
autorestart = true
redirect_stderr = true
startsecs = 3
directory = %(here)s/node0
command = cronosd start --home . --trace
stdout_logfile = %(here)s/node0.log

[program:cronos_777-1-node1]
autostart = true
autorestart = true
redirect_stderr = true
startsecs = 3
directory = %(here)s/node1
command = /Users/mavis/Desktop/authz/no_reg/cronosd start --home . --trace
stdout_logfile = %(here)s/node1.log

[program:cronos_777-1-node2]
autostart = true
autorestart = true
redirect_stderr = true
startsecs = 3
directory = %(here)s/node2
command = /Users/mavis/Desktop/authz/no_reg/cronosd start --home . --trace
stdout_logfile = %(here)s/node2.log
    """

    with open(cronos.base_dir / "tasks.ini", "w") as ini_file:
        ini_file.write(updated)

    cronos.supervisorctl("update")
    time.sleep(1)
    cli = cronos.cosmos_cli()
    wait_for_new_blocks(cli, 1)
    encoded_tx = "CvYBCroBCh4vY29zbW9zLmF1dGh6LnYxYmV0YTEuTXNnR3JhbnQSlwEKKmNyYzF4N3g5cGtmeGYzM2w4N2Z0c3BrNWFldHdua3IwbHZsdjMzNDZjZBIqY3JjMTZ6MGhlcno5OTg5NDZ3cjY1OWxyODRjOGM1NTZkYTU1ZGMzNGhoGj0KOwomL2Nvc21vcy5iYW5rLnYxYmV0YTEuU2VuZEF1dGhvcml6YXRpb24SEQoPCghiYXNldGNybxIDMjAw+j82Ci8vZXRoZXJtaW50LnR5cGVzLnYxLkV4dGVuc2lvbk9wdGlvbkR5bmFtaWNGZWVUeBIDCgEwEl8KVwpPCigvZXRoZXJtaW50LmNyeXB0by52MS5ldGhzZWNwMjU2azEuUHViS2V5EiMKIQNg/r8Tea2PYFq2XE07fpnN97ASePg32cuO4HmUkEGpFBIECgIIARIEEMCaDBpBqjIjjd67xL2FGigffzzTpmhj4tzdCCoUGe+SIU0/nR8c9yb44BDiK0pmId4z6IOfHR3383EvTa2b/Z+mj3UgJAE="
    res = requests.post("http://localhost:27007", json=json_rpc_send_body(encoded_tx))
    print("mm-res", res.json())
    time.sleep(1200)


def json_rpc_send_body(raw, method="broadcast_tx_async"):
    return {
        "jsonrpc": "2.0",
        "method": method,
        "params": {"tx": raw},
        "id": 1,
    }
