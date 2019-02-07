import cbor from 'cbor'
import autoAddMapDelimiters from 'test-helpers/autoAddMapDelimiters'

const decodeDietCBOR = data => {
  return cbor.decodeFirst(autoAddMapDelimiters(data))
}

export default decodeDietCBOR
