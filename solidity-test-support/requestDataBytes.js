import functionSelector from 'test-helpers/functionSelector'
import abiEncode from 'test-helpers/abiEncode'

const requestDataBytes = (specId, to, fHash, nonce, data) => {
  const types = ['address', 'uint256', 'uint256', 'bytes32', 'address', 'bytes4', 'uint256', 'bytes']
  const values = [0, 0, 1, specId, to, fHash, nonce, data]
  const encoded = abiEncode(types, values)
  const funcSelector = functionSelector('requestData(address,uint256,uint256,bytes32,address,bytes4,uint256,bytes)')
  return funcSelector + encoded
}

export default requestDataBytes
