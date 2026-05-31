'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { useCamera } from '@/hooks/useCamera'
import { parseQRToken } from '@/lib/qr'

type Tab = 'qr' | 'credential'

export default function LoginPage() {
  const router = useRouter()
  const [tab, setTab] = useState<Tab>('qr')
  const [businessId, setBusinessId] = useState('')
  const [operatorCode, setOperatorCode] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const scanInterval = useRef<ReturnType<typeof setInterval> | null>(null)
  const { videoRef, status, error: camError, start, stop } = useCamera()
  const { isAuthenticated, isLoading, loginQr, loginCredential } = useAuth()

  // Kalau sudah login, langsung ke order
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace('/order')
    }
  }, [isAuthenticated, isLoading, router])

  // Scan QR loop
  const startScan = useCallback(async () => {
    await start()
    scanInterval.current = setInterval(async () => {
      if (!videoRef.current || !canvasRef.current) return
      const video = videoRef.current
      const canvas = canvasRef.current
      const ctx = canvas.getContext('2d')
      if (!ctx || video.readyState < 2) return
      canvas.width = video.videoWidth
      canvas.height = video.videoHeight
      ctx.drawImage(video, 0, 0)

      // BarcodeDetector API (Chrome/Android)
      if ('BarcodeDetector' in window) {
        try {
          // @ts-expect-error BarcodeDetector
          const detector = new BarcodeDetector({ formats: ['qr_code'] })
          const codes = await detector.detect(canvas)
          if (codes.length > 0) {
            const raw = codes[0].rawValue as string
            const result = parseQRToken(raw)
            if (result.type === 'login') {
              handleQRLogin(result.text)
            }
          }
        } catch { /* silent */ }
      }
    }, 500)
  }, [start, videoRef])

  useEffect(() => {
    if (tab === 'qr') startScan()
    else {
      stop()
      if (scanInterval.current) clearInterval(scanInterval.current)
    }
    return () => {
      stop()
      if (scanInterval.current) clearInterval(scanInterval.current)
    }
  }, [tab, startScan, stop])

  const handleQRLogin = async (qr_token: string) => {
    if (loading) return
    setLoading(true)
    setError('')
    try {
      await loginQr(qr_token)
      router.replace('/order')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'QR tidak valid')
    } finally {
      setLoading(false)
    }
  }

  const handleCredentialLogin = async () => {
    if (!businessId || !operatorCode || !password) {
      setError('Semua field wajib diisi')
      return
    }
    setLoading(true)
    setError('')
    try {
      await loginCredential({ business_id: businessId, operator_code: operatorCode, password })
      router.replace('/order')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Login gagal')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', background: '#0d0e0f', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 20 }}>

      {/* Logo */}
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <div style={{ width: 56, height: 56, borderRadius: 16, background: '#ca400a', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 12px', fontSize: 24, fontWeight: 900, color: '#fff' }}>Q</div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#e3e2e2' }}>QIOS Kasir</div>
        <div style={{ fontSize: 13, color: '#8c8c8c', marginTop: 4 }}>Masuk untuk mulai transaksi</div>
      </div>

      {/* Card */}
      <div style={{ width: '100%', maxWidth: 380, background: '#121414', borderRadius: 20, border: '0.5px solid #343535', overflow: 'hidden' }}>

        {/* Tabs */}
        <div style={{ display: 'flex', borderBottom: '0.5px solid #343535' }}>
          {(['qr', 'credential'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => { setTab(t); setError('') }}
              style={{ flex: 1, padding: '14px', fontSize: 13, fontWeight: 700, border: 'none', cursor: 'pointer', background: tab === t ? '#1b1c1c' : 'transparent', color: tab === t ? '#ca400a' : '#8c8c8c', borderBottom: tab === t ? '2px solid #ca400a' : '2px solid transparent', transition: 'all .2s' }}
            >
              {t === 'qr' ? '📷 Scan QR' : '⌨️ Kode Kasir'}
            </button>
          ))}
        </div>

        <div style={{ padding: 20 }}>

          {/* QR Tab */}
          {tab === 'qr' && (
            <div>
              <div style={{ background: '#0d0e0f', borderRadius: 12, overflow: 'hidden', aspectRatio: '1', position: 'relative', marginBottom: 16 }}>
                <video ref={videoRef} style={{ width: '100%', height: '100%', objectFit: 'cover' }} playsInline muted />
                <canvas ref={canvasRef} style={{ display: 'none' }} />
                {status === 'requesting' && (
                  <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0d0e0f', color: '#8c8c8c', fontSize: 13 }}>
                    Meminta akses kamera...
                  </div>
                )}
                {status === 'idle' && (
                  <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0d0e0f', flexDirection: 'column', gap: 12 }}>
                    <div style={{ fontSize: 40 }}>📷</div>
                    <div style={{ color: '#8c8c8c', fontSize: 13 }}>Kamera belum aktif</div>
                  </div>
                )}
                {status === 'active' && (
                  <div style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }}>
                    <div style={{ position: 'absolute', top: '25%', left: '25%', right: '25%', bottom: '25%', border: '2px solid #ca400a', borderRadius: 12 }} />
                  </div>
                )}
              </div>
              {camError && <div style={{ color: '#ef4444', fontSize: 12, textAlign: 'center', marginBottom: 12 }}>{camError}</div>}
              <div style={{ fontSize: 12, color: '#8c8c8c', textAlign: 'center' }}>
                Arahkan kamera ke QR code dari dashboard owner
              </div>
            </div>
          )}

          {/* Credential Tab */}
          {tab === 'credential' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <label style={{ fontSize: 11, color: '#8c8c8c', fontWeight: 600, display: 'block', marginBottom: 6 }}>Business ID</label>
                <input
                  value={businessId}
                  onChange={e => setBusinessId(e.target.value)}
                  placeholder="00000000-0000-0000-0000-000000000000"
                  style={{ width: '100%', padding: '10px 12px', background: '#0d0e0f', border: '0.5px solid #343535', borderRadius: 10, color: '#e3e2e2', fontSize: 13, outline: 'none', fontFamily: 'Inter, sans-serif' }}
                />
              </div>
              <div>
                <label style={{ fontSize: 11, color: '#8c8c8c', fontWeight: 600, display: 'block', marginBottom: 6 }}>Operator Code</label>
                <input
                  value={operatorCode}
                  onChange={e => setOperatorCode(e.target.value)}
                  placeholder="kasir-1"
                  style={{ width: '100%', padding: '10px 12px', background: '#0d0e0f', border: '0.5px solid #343535', borderRadius: 10, color: '#e3e2e2', fontSize: 13, outline: 'none', fontFamily: 'Inter, sans-serif' }}
                />
              </div>
              <div>
                <label style={{ fontSize: 11, color: '#8c8c8c', fontWeight: 600, display: 'block', marginBottom: 6 }}>Password</label>
                <input
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  placeholder="••••••••"
                  style={{ width: '100%', padding: '10px 12px', background: '#0d0e0f', border: '0.5px solid #343535', borderRadius: 10, color: '#e3e2e2', fontSize: 13, outline: 'none', fontFamily: 'Inter, sans-serif' }}
                />
              </div>
              <button
                onClick={handleCredentialLogin}
                disabled={loading}
                style={{ width: '100%', padding: '13px', background: loading ? '#343535' : '#ca400a', color: '#fff', border: 'none', borderRadius: 10, fontSize: 14, fontWeight: 700, cursor: loading ? 'not-allowed' : 'pointer', marginTop: 4 }}
              >
                {loading ? 'Memproses...' : 'Masuk Kasir →'}
              </button>
            </div>
          )}

          {error && (
            <div style={{ marginTop: 12, padding: '10px 12px', background: 'rgba(239,68,68,0.1)', border: '0.5px solid rgba(239,68,68,0.3)', borderRadius: 8, color: '#ef4444', fontSize: 13, textAlign: 'center' }}>
              {error}
            </div>
          )}
        </div>
      </div>

      <div style={{ marginTop: 20, fontSize: 12, color: '#343535' }}>QIOS · PT Skalar Solusi Digital</div>
    </div>
  )
}