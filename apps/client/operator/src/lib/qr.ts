export type QRResult = {
  text: string
  type: 'login' | 'unknown'
}

export function parseQRToken(raw: string): QRResult {
  try {
    if (raw.startsWith('qios:op:')) {
      return { text: raw.replace('qios:op:', ''), type: 'login' }
    }
    const parsed = JSON.parse(raw)
    if (parsed?.token) {
      return { text: parsed.token, type: 'login' }
    }
  } catch {
    // not JSON
  }
  return { text: raw, type: 'unknown' }
}