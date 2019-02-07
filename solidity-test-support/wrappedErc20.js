import BN from 'bn.js'

const bNToStringOrIdentity = a => BN.isBN(a) ? a.toString() : a

// Deal with transfer amount type truffle doesn't currently handle. (BN)
const wrappedERC20 = contract => ({
  ...contract,
  transfer: async (address, amount) =>
    contract.transfer(address, bNToStringOrIdentity(amount)),
  transferAndCall: async (address, amount, payload, options) =>
    contract.transferAndCall(address, bNToStringOrIdentity(amount), payload, options)
})

export default wrappedERC20
