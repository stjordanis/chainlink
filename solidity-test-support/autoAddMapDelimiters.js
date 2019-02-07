const startMapBuffer = Buffer.from([0xBF])
const endMapBuffer = Buffer.from([0xFF])

const autoAddMapDelimiters = data => {
  let buffer = data

  if (buffer[0] >> 5 !== 5) {
    buffer = Buffer.concat([startMapBuffer, buffer, endMapBuffer], buffer.length + 2)
  }

  return buffer
}

export default autoAddMapDelimiters
