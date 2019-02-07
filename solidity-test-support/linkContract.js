import wrappedERC20 from 'test-helpers/wrappedERC20'
import Utils from 'test-helpers/utils'
import Wallet from 'test-helpers/wallet'
import Deployer from 'test-helpers/deployer'

const PRIVATE_KEY = 'c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3'
const utils = Utils(web3.currentProvider)
const wallet = Wallet(PRIVATE_KEY, utils)
const deployer = Deployer(wallet, utils)

const deploy = async (filePath, ...args) => deployer.perform(filePath, ...args)

const linkContract = async () => wrappedERC20(await deploy('link_token/contracts/LinkToken.sol'))

export default linkContract
