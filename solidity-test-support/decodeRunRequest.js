import abi from 'ethereumjs-abi'
import util from 'ethereumjs-util'
import toHex from 'test-helpers/toHex'
import hexToAddress from 'test-helpers/hexToAddress'
import zeroX from 'test-helpers/zeroX'
import autoAddMapDelimiters from 'test-helpers/autoAddMapDelimiters'

const decodeRunRequest = log => {
  const runABI = util.toBuffer(log.data)
  const types = ['uint256', 'uint256', 'address', 'bytes4', 'uint256', 'bytes']
  const [
    requestId,
    version,
    callbackAddress,
    callbackFunc,
    expiration,
    data
  ] = abi.rawDecode(types, runABI)

  return {
    topic: log.topics[0],
    jobId: log.topics[1],
    requester: hexToAddress(log.topics[2]),
    payment: log.topics[3],
    id: toHex(requestId),
    dataVersion: version,
    callbackAddr: zeroX(callbackAddress),
    callbackFunc: toHex(callbackFunc),
    expiration: toHex(expiration),
    data: autoAddMapDelimiters(data)
  }
}

export default decodeRunRequest
