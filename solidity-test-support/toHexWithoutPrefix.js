import BN from 'bn.js'

const toHexWithoutPrefix = arg => {
  if (arg instanceof Buffer || arg instanceof BN) {
    return arg.toString('hex')
  } else if (arg instanceof Uint8Array) {
    return Array.prototype.reduce.call(arg, (a, v) => a + v.toString('16').padStart(2, '0'), '')
  } else {
    return Buffer.from(arg, 'ascii').toString('hex')
  }
}

export default toHexWithoutPrefix
