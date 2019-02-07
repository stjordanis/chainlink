import zeroX from 'test-helpers/zeroX'
import toHexWithoutPrefix from 'test-helpers/toHexWithoutPrefix'

const toHex = value => {
  return zeroX(toHexWithoutPrefix(value))
}

export default toHex
