const INVALIDVALUE = {
  // If you got this value, you probably tried to use one of the variables below
  // before they were initialized. Do any test initialization which requires
  // them in a callback passed to Mocha's `before` or `beforeEach`.
  // https://mochajs.org/#asynchronous-hooks
  unitializedValueProbablyShouldUseVaribleInMochaBeforeCallback: null
}

const [oracleNode] = Array(1000).fill(INVALIDVALUE)

export default oracleNode
