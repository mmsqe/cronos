import base64
import hashlib
import json

import pytest

from .ibc_utils import (
    RATIO,
    assert_ready,
    get_balance,
    hermes_transfer,
    prepare_network,
)
from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
    eth_to_bech32,
    parse_events,
    parse_events_rpc,
    send_transaction,
    setup_token_mapping,
    wait_for_fn,
    wait_for_new_blocks,
)


@pytest.fixture(scope="module", params=[True, False])
def ibc(request, tmp_path_factory):
    "prepare-network"
    incentivized = request.param
    name = "ibc"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, incentivized)
    yield from network


def get_balances(chain, addr):
    return chain.cosmos_cli().balances(addr)


def find_duplicate(attributes):
    res = set()
    key = attributes[0]["key"]
    for attribute in attributes:
        if attribute["key"] == key:
            value0 = attribute["value"]
        elif attribute["key"] == "amount":
            amount = attribute["value"]
            value_pair = f"{value0}:{amount}"
            if value_pair in res:
                return value_pair
            res.add(value_pair)
    return None


def test_ibc_transfer_with_hermes(ibc):
    """
    test ibc transfer tokens with hermes cli
    """
    src_amount = hermes_transfer(ibc)
    dst_amount = src_amount * RATIO  # the decimal places difference
    dst_denom = "basetcro"
    dst_addr = eth_to_bech32(ADDRS["signer2"])
    old_dst_balance = get_balance(ibc.cronos, dst_addr, dst_denom)

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.cronos, dst_addr, dst_denom)
        return new_dst_balance != old_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + dst_amount == new_dst_balance

    # assert that the relayer transactions do enables the dynamic fee extension option.
    cli = ibc.cronos.cosmos_cli()
    criteria = "message.action=/ibc.core.channel.v1.MsgChannelOpenInit"
    tx = cli.tx_search(criteria)["txs"][0]
    events = parse_events_rpc(tx["events"])
    fee = int(events["tx"]["fee"].removesuffix("basetcro"))
    gas = int(tx["gas_wanted"])
    # the effective fee is decided by the max_priority_fee (base fee is zero)
    # rather than the normal gas price
    assert fee == gas * 1000000

    # check duplicate OnRecvPacket events
    criteria = "message.action=/ibc.core.channel.v1.MsgRecvPacket"
    tx = cli.tx_search(criteria)["txs"][0]
    events = tx["logs"][1]["events"]
    for event in events:
        dup = find_duplicate(event["attributes"])
        assert not dup, f"duplicate {dup} in {event['type']}"


def test_ibc_incentivized_transfer(ibc):
    if not ibc.incentivized:
        # this test case only works for incentivized channel.
        return
    src_chain = ibc.cronos.cosmos_cli()
    dst_chain = ibc.chainmain.cosmos_cli()
    receiver = dst_chain.address("signer2")
    sender = src_chain.address("signer2")
    relayer = src_chain.address("signer1")
    original_amount = src_chain.balance(relayer, denom="ibcfee")
    original_amount_sender = src_chain.balance(sender, denom="ibcfee")

    rsp = src_chain.ibc_transfer(
        sender,
        receiver,
        "1000basetcro",
        "channel-0",
        1,
        "100000000basecro",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    evt = parse_events(rsp["logs"])["send_packet"]
    print("packet event", evt)
    packet_seq = int(evt["packet_sequence"])

    rsp = src_chain.pay_packet_fee(
        "transfer",
        "channel-0",
        packet_seq,
        recv_fee="10ibcfee",
        ack_fee="10ibcfee",
        timeout_fee="10ibcfee",
        from_=sender,
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # fee is locked
    assert src_chain.balance(sender, denom="ibcfee") == original_amount_sender - 30

    # wait for relayer receive the fee
    def check_fee():
        amount = src_chain.balance(relayer, denom="ibcfee")
        if amount > original_amount:
            assert amount == original_amount + 20
            return True
        else:
            return False

    wait_for_fn("wait for relayer to receive the fee", check_fee)

    # timeout fee is refunded
    assert src_chain.balance(sender, denom="ibcfee") == original_amount_sender - 20


def test_cronos_transfer_tokens(ibc):
    """
    test sending basetcro from cronos to crypto-org-chain using cli transfer_tokens.
    depends on `test_ibc` to send the original coins.
    """
    assert_ready(ibc)
    dst_addr = ibc.chainmain.cosmos_cli().address("signer2")
    dst_amount = 2
    dst_denom = "basecro"
    cli = ibc.cronos.cosmos_cli()
    src_amount = dst_amount * RATIO  # the decimal places difference
    src_addr = cli.address("signer2")
    src_denom = "basetcro"

    # case 1: use cronos cli
    old_src_balance = get_balance(ibc.cronos, src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)
    rsp = cli.transfer_tokens(
        src_addr,
        dst_addr,
        f"{src_amount}{src_denom}",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + dst_amount == new_dst_balance
    new_src_balance = get_balance(ibc.cronos, src_addr, src_denom)
    assert old_src_balance - src_amount == new_src_balance


def test_cronos_transfer_tokens_acknowledgement_error(ibc):
    """
    test sending basetcro from cronos to crypto-org-chain using cli transfer_tokens
    with invalid receiver for acknowledgement error.
    depends on `test_ibc` to send the original coins.
    """
    assert_ready(ibc)
    dst_addr = "invalid_address"
    dst_amount = 2
    cli = ibc.cronos.cosmos_cli()
    src_amount = dst_amount * RATIO  # the decimal places difference
    src_addr = cli.address("signer2")
    src_denom = "basetcro"

    old_src_balance = get_balance(ibc.cronos, src_addr, src_denom)
    rsp = cli.transfer_tokens(
        src_addr,
        dst_addr,
        f"{src_amount}{src_denom}",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_src_balance = 0

    def check_balance_change():
        nonlocal new_src_balance
        new_src_balance = get_balance(ibc.cronos, src_addr, src_denom)
        return old_src_balance == new_src_balance

    wait_for_fn("balance no change", check_balance_change)
    new_src_balance = get_balance(ibc.cronos, src_addr, src_denom)


def test_cro_bridge_contract(ibc):
    """
    test sending basetcro from cronos to crypto-org-chain using CroBridge contract.
    depends on `test_ibc` to send the original coins.
    """
    dst_addr = ibc.chainmain.cosmos_cli().address("signer2")
    dst_amount = 2
    dst_denom = "basecro"
    src_amount = dst_amount * RATIO  # the decimal places difference
    old_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)

    # case 2: use CroBridge contract
    w3 = ibc.cronos.w3
    contract = deploy_contract(w3, CONTRACTS["CroBridge"])
    tx = contract.functions.send_cro_to_crypto_org(dst_addr).build_transaction(
        {"from": ADDRS["signer2"], "value": src_amount}
    )
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)
        return old_dst_balance != new_dst_balance

    wait_for_fn("check balance change", check_balance_change)
    assert old_dst_balance + dst_amount == new_dst_balance


def test_cronos_transfer_source_tokens(ibc):
    """
    test sending crc20 tokens originated from cronos to crypto-org-chain
    """
    assert_ready(ibc)
    # deploy crc21 contract
    w3 = ibc.cronos.w3
    contract, denom = setup_token_mapping(ibc.cronos, "TestERC21Source", "DOG")
    # send token to crypto.org
    print("send to crypto.org")
    chainmain_receiver = ibc.chainmain.cosmos_cli().address("signer2")
    dest_denom = ibc_denom("channel-0", denom)
    amount = 1000

    # check and record receiver balance
    chainmain_receiver_balance = get_balance(
        ibc.chainmain, chainmain_receiver, dest_denom
    )
    assert chainmain_receiver_balance == 0

    # send to ibc
    tx = contract.functions.send_to_ibc_v2(
        chainmain_receiver, amount, 0, b""
    ).build_transaction({"from": ADDRS["validator"]})
    txreceipt = send_transaction(w3, tx)
    assert txreceipt.status == 1, "should success"

    # check balance
    chainmain_receiver_new_balance = 0

    def check_chainmain_balance_change():
        nonlocal chainmain_receiver_new_balance
        chainmain_receiver_new_balance = get_balance(
            ibc.chainmain, chainmain_receiver, dest_denom
        )
        chainmain_receiver_all_balance = get_balances(ibc.chainmain, chainmain_receiver)
        print("receiver all balance:", chainmain_receiver_all_balance)
        return chainmain_receiver_balance != chainmain_receiver_new_balance

    wait_for_fn("check balance change", check_chainmain_balance_change)
    assert chainmain_receiver_new_balance == amount

    # check legacy send to ibc
    tx = contract.functions.send_to_ibc(chainmain_receiver, 1).build_transaction(
        {"from": ADDRS["validator"]}
    )
    txreceipt = send_transaction(w3, tx)
    assert txreceipt.status == 0, "should fail"

    # send back the token to cronos
    # check receiver balance
    cronos_balance_before_send = contract.caller.balanceOf(ADDRS["signer2"])
    assert cronos_balance_before_send == 0

    # send back token through ibc
    print("Send back token through ibc")
    chainmain_cli = ibc.chainmain.cosmos_cli()
    cronos_receiver = eth_to_bech32(ADDRS["signer2"])

    coin = "1000" + dest_denom
    rsp = chainmain_cli.ibc_transfer(
        chainmain_receiver, cronos_receiver, coin, "channel-0", 1, "100000000basecro"
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # check contract balance
    cronos_balance_after_send = 0

    def check_contract_balance_change():
        nonlocal cronos_balance_after_send
        cronos_balance_after_send = contract.caller.balanceOf(ADDRS["signer2"])
        return cronos_balance_after_send != cronos_balance_before_send

    wait_for_fn("check contract balance change", check_contract_balance_change)
    assert cronos_balance_after_send == amount


def test_cronos_transfer_source_tokens_with_proxy(ibc):
    """
    test sending crc20 tokens originated from cronos to crypto-org-chain
    """
    assert_ready(ibc)
    # deploy crc20 contract
    w3 = ibc.cronos.w3
    symbol = "TEST"
    contract, denom = setup_token_mapping(ibc.cronos, "TestCRC20", symbol)

    # deploy crc20 proxy contract
    proxycrc20 = deploy_contract(
        w3,
        CONTRACTS["TestCRC20Proxy"],
        (contract.address, True),
    )

    print("proxycrc20 contract deployed at address: ", proxycrc20.address)
    assert proxycrc20.caller.is_source()
    assert proxycrc20.caller.crc20() == contract.address

    cronos_cli = ibc.cronos.cosmos_cli()
    # change token mapping
    rsp = cronos_cli.update_token_mapping(
        denom, proxycrc20.address, symbol, 6, from_="validator"
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(cronos_cli, 1)

    print("check the contract mapping exists now")
    rsp = cronos_cli.query_denom_by_contract(proxycrc20.address)
    assert rsp["denom"] == denom

    # send token to crypto.org
    print("send to crypto.org")
    chainmain_receiver = ibc.chainmain.cosmos_cli().address("signer2")
    dest_denom = ibc_denom("channel-0", denom)
    amount = 1000
    sender = ADDRS["validator"]

    # First we need to approve the proxy contract to move asset
    tx = contract.functions.approve(proxycrc20.address, amount).build_transaction(
        {"from": sender}
    )
    txreceipt = send_transaction(w3, tx)
    assert txreceipt.status == 1, "should success"
    assert contract.caller.allowance(ADDRS["validator"], proxycrc20.address) == amount

    # check and record receiver balance
    chainmain_receiver_balance = get_balance(
        ibc.chainmain, chainmain_receiver, dest_denom
    )
    assert chainmain_receiver_balance == 0

    # send to ibc
    tx = proxycrc20.functions.send_to_ibc(
        chainmain_receiver, amount, 0, b""
    ).build_transaction({"from": sender})
    txreceipt = send_transaction(w3, tx)
    print(txreceipt)
    assert txreceipt.status == 1, "should success"

    # check balance
    chainmain_receiver_new_balance = 0

    def check_chainmain_balance_change():
        nonlocal chainmain_receiver_new_balance
        chainmain_receiver_new_balance = get_balance(
            ibc.chainmain, chainmain_receiver, dest_denom
        )
        chainmain_receiver_all_balance = get_balances(ibc.chainmain, chainmain_receiver)
        print("receiver all balance:", chainmain_receiver_all_balance)
        return chainmain_receiver_balance != chainmain_receiver_new_balance

    wait_for_fn("check balance change", check_chainmain_balance_change)
    assert chainmain_receiver_new_balance == amount

    # send back the token to cronos
    # check receiver balance
    cronos_balance_before_send = contract.caller.balanceOf(ADDRS["signer2"])
    assert cronos_balance_before_send == 0

    # send back token through ibc
    print("Send back token through ibc")
    chainmain_cli = ibc.chainmain.cosmos_cli()
    cronos_receiver = eth_to_bech32(ADDRS["signer2"])

    coin = "1000" + dest_denom
    rsp = chainmain_cli.ibc_transfer(
        chainmain_receiver, cronos_receiver, coin, "channel-0", 1, "100000000basecro"
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # check contract balance
    cronos_balance_after_send = 0

    def check_contract_balance_change():
        nonlocal cronos_balance_after_send
        cronos_balance_after_send = contract.caller.balanceOf(ADDRS["signer2"])
        return cronos_balance_after_send != cronos_balance_before_send

    wait_for_fn("check contract balance change", check_contract_balance_change)
    assert cronos_balance_after_send == amount


def ibc_denom(channel, denom):
    h = hashlib.sha256(f"transfer/{channel}/{denom}".encode()).hexdigest().upper()
    return f"ibc/{h}"


def test_ica(ibc, tmp_path):
    connid = "connection-0"
    cli_host = ibc.chainmain.cosmos_cli()
    cli_controller = ibc.cronos.cosmos_cli()

    print("register ica account")
    rsp = cli_controller.ica_register_account(
        connid, from_="signer2", gas="400000", fees="100000000basetcro"
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    port_id, channel_id = next(
        (
            base64.b64decode(evt["attributes"][0]["value"].encode()).decode(),
            base64.b64decode(evt["attributes"][1]["value"].encode()).decode(),
        )
        for evt in rsp["events"]
        if evt["type"] == "channel_open_init"
    )
    print("port-id", port_id, "channel-id", channel_id)

    print("wait for ica channel ready")

    def check_channel_ready():
        channels = cli_controller.ibc_query_channels(connid)["channels"]
        try:
            state = next(
                channel["state"]
                for channel in channels
                if channel["channel_id"] == channel_id
            )
        except StopIteration:
            return False
        return state == "STATE_OPEN"

    wait_for_fn("channel ready", check_channel_ready)

    print("query ica account")
    ica_address = cli_controller.ica_query_account(
        connid, cli_controller.address("signer2")
    )["interchainAccountAddress"]
    print("ica address", ica_address)

    # initial balance of interchain account should be zero
    assert cli_host.balance(ica_address) == 0

    # send some funds to interchain account
    rsp = cli_host.transfer("signer2", ica_address, "1cro", gas_prices="1000000basecro")
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(cli_host, 1)

    # check if the funds are received in interchain account
    assert cli_host.balance(ica_address, denom="basecro") == 100000000

    # generate a transaction to send to host chain
    generated_tx = tmp_path / "generated_tx.txt"
    generated_tx_msg = {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": ica_address,
        "to_address": cli_host.address("signer2"),
        "amount": [{"denom": "basecro", "amount": "50000000"}],
    }
    str = json.dumps(generated_tx_msg)
    generated_packet = cli_controller.ica_generate_packet_data(str)
    print(generated_packet)
    generated_tx.write_text(json.dumps(generated_packet))

    num_txs = len(cli_host.query_all_txs(ica_address)["txs"])

    # submit transaction on host chain on behalf of interchain account
    rsp = cli_controller.ica_submit_tx(
        connid,
        generated_tx,
        from_="signer2",
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    packet_seq = next(
        int(base64.b64decode(evt["attributes"][4]["value"].encode()))
        for evt in rsp["events"]
        if evt["type"] == "send_packet"
    )
    print("packet sequence", packet_seq)

    def check_ica_tx():
        return len(cli_host.query_all_txs(ica_address)["txs"]) > num_txs

    print("wait for ica tx arrive")
    wait_for_fn("ica transfer tx", check_ica_tx)

    # check if the funds are reduced in interchain account
    assert cli_host.balance(ica_address, denom="basecro") == 50000000
