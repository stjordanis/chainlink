const fulfillOracleRequest = async (oracle, request, response, options) => {
  if (!options) options = {}

  return oracle.fulfillData(
    request.id,
    request.payment,
    request.callbackAddr,
    request.callbackFunc,
    request.expiration,
    response,
    options)
}

export default fulfillOracleRequest
