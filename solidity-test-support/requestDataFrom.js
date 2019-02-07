// link param must be from linkContract(), if amount is a BN
const requestDataFrom = (oc, link, amount, args, options) => {
  if (!options) options = {}
  return link.transferAndCall(oc.address, amount, args, options)
}

export default requestDataFrom
