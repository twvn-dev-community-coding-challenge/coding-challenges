/** @param {number} len */
export function randomDigits(len) {
  const buf = new Uint32Array(len)
  crypto.getRandomValues(buf)
  let s = ""
  for (let i = 0; i < len; i += 1) {
    s += String(buf[i] % 10)
  }
  return s
}

export function generateDemoUsername() {
  return `user_${Date.now().toString(36)}_${randomDigits(4)}`
}

export function generateDemoPassword() {
  const pool = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
  const buf = new Uint32Array(14)
  crypto.getRandomValues(buf)
  let s = ""
  for (let i = 0; i < buf.length; i += 1) {
    s += pool[buf[i] % pool.length]
  }
  return `Aa1!${s}`
}

/**
 * National-style digits with trunk 0 (as users type locally). Backend normalizes to E.164.
 * @param {string} countryCode ISO2: PH | VN | TH | SG
 */
export function generateDemoPhone(countryCode) {
  switch (countryCode) {
    case "PH":
      return `0917${randomDigits(7)}`
    case "VN":
      return `09${randomDigits(8)}`
    case "TH":
      return `08${randomDigits(8)}`
    case "SG":
      return `08${randomDigits(7)}`
    default:
      return `0917${randomDigits(7)}`
  }
}
