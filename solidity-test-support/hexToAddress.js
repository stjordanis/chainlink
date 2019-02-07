import zeroX from 'test-helpers/zeroX'
import bigNum from 'test-helpers/bigNum'

const hexToAddress = hex => zeroX(bigNum(hex).toString('hex'))

export default hexToAddress
