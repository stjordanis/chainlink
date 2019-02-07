export default signature =>
  '0x' + web3.utils.sha3(signature).slice(2).slice(0, 8)
