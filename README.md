## A Modification of Binance Smart Chain client v1.1.9
### (This version is outdated (prior The Merge) and does not work properly compared to the latest version of BSC client)

The goal of Binance Smart Chain is to bring programmability and interoperability to Binance Chain. In order to embrace the existing popular community and advanced technology, it will bring huge benefits by staying compatible with all the existing smart contracts on Ethereum and Ethereum tooling. And to achieve that, the easiest solution is to develop based on go-ethereum fork, as we respect the great work of Ethereum very much.

Binance Smart Chain starts its development based on go-ethereum fork. So you may see many toolings, binaries and also docs are based on Ethereum ones, such as the name “geth”.

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/ethereum/go-ethereum?tab=doc)
[![Discord](https://img.shields.io/badge/discord-join%20chat-blue.svg)](https://discord.gg/z2VpC455eU)

But from that baseline of EVM compatible, Binance Smart Chain introduces  a system of 21 validators with Proof of Staked Authority (PoSA) consensus that can support short block time and lower fees. The most bonded validator candidates of staking will become validators and produce blocks. The double-sign detection and other slashing logic guarantee security, stability, and chain finality.

Cross-chain transfer and other communication are possible due to native support of interoperability. Relayers and on-chain contracts are developed to support that. Binance DEX remains a liquid venue of the exchange of assets on both chains. This dual-chain architecture will be ideal for users to take advantage of the fast trading on one side and build their decentralized apps on the other side. **The Binance Smart Chain** will be:

- **A self-sovereign blockchain**: Provides security and safety with elected validators.
- **EVM-compatible**: Supports all the existing Ethereum tooling along with faster finality and cheaper transaction fees.
- **Interoperable**: Comes with efficient native dual chain communication; Optimized for scaling high-performance dApps that require fast and smooth user experience.
- **Distributed with on-chain governance**: Proof of Staked Authority brings in decentralization and community participants. As the native token, BNB will serve as both the gas of smart contract execution and tokens for staking.

More details in [White Paper](http://binance.org/en#smartChain).

## Functionalities

Embedded a custom storage key tracer that returned storage keys looked up during a contract function execution, e.g., fetching the storage key for reserve values for an LP. Tracers are executed by pointing RPC calls to the debug namespace.

Ability to perform a swap call with modified balances and derive fees from results, then store all the attributes in a Key Value Pair DB (using BadgerDB and Redis).

Ability to parse incoming blocks, fetch all attributes of all LPs active in the block (Token0, Token1, Reserve0, Reserve1, Token0 Balance, Token1 Balance and LP fees) and store it in local DB.

Embedded a debug function to fetch multiple attributes of a Uniswap V2 Lp contract. “debug_fetchSloadValues” can be simply called via RPC.

Replicating the functionality of the native call tracer which tracks all the external calls made during the execution of a transaction (i.e. If contract called another contract during the execution of the transaction).

Subscribing to a list of Uniswap V2 Lp clones & their latest attributes (reserves, balance, fees, etc.) extracted from new & old headers.

Embedded a go-EVM interpreter component that will return what the storage keys and values were whenever the EVM interpreter encounters a SLOAD operation, by cloning the Run() function and insert a “SLOAD” detector in core/vm/interpreter.go; and cloning run() and Call() function in core/vm/evm.go. Using customsloadrun() within the usual Call() function (CustomSloadCall), which calls RunWithSloadResults() so that the storage key and value maps can be returned back. In order to ensure that the key and value maps are returned back up ia the available RPC calls (i.e. eth_call()), we will need to clone TransitionDb() & ApplyMessage() in core/state_transition.go.

## License

The bsc library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The bsc binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
