import abi from 'ethereumjs-abi'

const abiEncode = (types, values) => {
  return abi.rawEncode(types, values).toString('hex')
}

export default abiEncode
