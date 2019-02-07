const assertActionThrows = action => (
  Promise
    .resolve()
    .then(action)
    .catch(error => {
      assert(error, 'Expected an error to be raised')
      assert(error.message, 'Expected an error to be raised')
      return error.message
    })
    .then(errorMessage => {
      assert(errorMessage, 'Expected an error to be raised')
      const invalidOpcode = errorMessage.includes('invalid opcode')
      const reverted = errorMessage.includes(
        'VM Exception while processing transaction: revert')
      assert(invalidOpcode || reverted,
        'expected following error message to include "invalid JUMP" or ' +
        `"revert": "${errorMessage}"`)
      // see https://github.com/ethereumjs/testrpc/issues/39
      // for why the "invalid JUMP" is the throw related error when using TestRPC
    })
)

export default assertActionThrows
