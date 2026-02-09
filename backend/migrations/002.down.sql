drop table if exists balance_snapshots;
drop view if exists balances;
drop view if exists asset_flows;

drop index if exists erc20_transfers_chain_id_token_address_from_address;
drop index if exists erc20_transfers_chain_id_token_address_to_address;
drop index if exists native_transfers_chain_id_from_address;
drop index if exists native_transfers_chain_id_to_address;