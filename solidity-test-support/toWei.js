import bigNum from 'test-helpers/bigNum'

const toWei = number => bigNum(web3.utils.toWei(bigNum(number)))

export default toWei
