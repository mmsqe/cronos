import time
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
    spend_limit = 200
    granter_address = cli.address("community")
    grantee_address = cli.address("signer1")

    # tx = cli.grant_authorization(
    #     grantee_address,
    #     "send",
    #     granter_address,
    #     spend_limit="%s%s" % (spend_limit, "basetcro"),
    #     generate_only=True,
    # )
    # tx = cli.transfer(
    #     granter_address,
    #     grantee_address,
    #     "1000basetcro",
    #     generate_only=True,
    # )
    # print("mm-tx", tx)
    # signed = cli.sign_tx_json(tx, granter_address)
    signed = {
        "body": {
            "messages": [
                {
                    "@type": "/cosmos.authz.v1beta1.MsgGrant",
                    "granter": "crc1x7x9pkfxf33l87ftspk5aetwnkr0lvlv3346cd",
                    "grantee": "crc16z0herz998946wr659lr84c8c556da55dc34hh",
                    "grant": {
                        "authorization": {
                            "@type": "/cosmos.bank.v1beta1.SendAuthorization",
                            "spend_limit": [
                                {
                                    "denom": "basetcro",
                                    "amount": "200"
                                }
                            ],
                            "allow_list": []
                        },
                        "expiration": None
                    }
                }
            ],
            "memo": "",
            "timeout_height": "0",
            "extension_options": [
                {
                    "@type": "/ethermint.types.v1.ExtensionOptionDynamicFeeTx",
                    "max_priority_price": "0"
                }
            ],
            "non_critical_extension_options": []
        },
        "auth_info": {
            "signer_infos": [
                {
                    "public_key": {
                        "@type": "/ethermint.crypto.v1.ethsecp256k1.PubKey",
                        "key": "A2D+vxN5rY9gWrZcTTt+mc33sBJ4+DfZy47geZSQQakU"
                    },
                    "mode_info": {
                        "single": {
                            "mode": "SIGN_MODE_DIRECT"
                        }
                    },
                    "sequence": "0"
                }
            ],
            "fee": {
                "amount": [],
                "gas_limit": "200000",
                "payer": "",
                "granter": ""
            },
            "tip": None
        },
        "signatures": [
            "qjIjjd67xL2FGigffzzTpmhj4tzdCCoUGe+SIU0/nR8c9yb44BDiK0pmId4z6IOfHR3383EvTa2b/Z+mj3UgJAE="
        ]
    }
    print("mm-signed", signed)
    rsp = cli.broadcast_tx_json(signed)
    # assert rsp["code"] == 0, rsp["raw_log"]
    print("mm-rsp", rsp)
    time.sleep(1200)
