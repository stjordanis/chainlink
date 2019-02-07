const zeroX = value => (value.slice(0, 2) !== '0x') ? `0x${value}` : value

export default zeroX
