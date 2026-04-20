import { useState } from "react"
import { Button } from "@/components/ui/button"
import { apiPost } from "@/lib/api"
import {
  generateDemoPassword,
  generateDemoPhone,
  generateDemoUsername,
} from "@/lib/demoRegisterValues"

const inputClass =
  "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"

const COUNTRIES = [
  { code: "PH", label: "Philippines (PH)" },
  { code: "VN", label: "Vietnam (VN)" },
  { code: "TH", label: "Thailand (TH)" },
  { code: "SG", label: "Singapore (SG)" },
]

export default function RegisterPage() {
  const [step, setStep] = useState(1)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [phoneNumber, setPhoneNumber] = useState("")
  const [country, setCountry] = useState("PH")
  const [otp, setOtp] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [registeredSummary, setRegisteredSummary] = useState(null)

  async function handleRegisterSubmit(e) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const data = await apiPost("/api/register", {
        username,
        password,
        phone_number: phoneNumber,
        country,
      })
      setRegisteredSummary(data.user ?? { username, phone_number: phoneNumber, status: "pending" })
      setStep(2)
      setOtp("")
    } catch (err) {
      setError(err.message || "Registration failed")
    } finally {
      setLoading(false)
    }
  }

  async function handleOtpSubmit(e) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await apiPost("/api/verify", {
        username,
        otp_code: otp.replace(/\D/g, "").slice(0, 6),
      })
      setStep(3)
    } catch (err) {
      setError(err.message || "Invalid code")
    } finally {
      setLoading(false)
    }
  }

  function resetFlow() {
    setStep(1)
    setUsername("")
    setPassword("")
    setPhoneNumber("")
    setCountry("PH")
    setOtp("")
    setError(null)
    setRegisteredSummary(null)
  }

  function fillRandomCredentials() {
    setError(null)
    setUsername(generateDemoUsername())
    setPassword(generateDemoPassword())
    setPhoneNumber(generateDemoPhone(country))
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-50 flex flex-col items-center justify-center p-6 font-sans">
      <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900/60 p-8 shadow-2xl backdrop-blur-md">
        <h1 className="text-2xl font-semibold tracking-tight text-white mb-1">
          Register new account
        </h1>

        {error ? (
          <div
            className="mb-4 rounded-lg border border-red-900/60 bg-red-950/40 px-3 py-2 text-sm text-red-200"
            role="alert"
          >
            {error}
          </div>
        ) : null}

        {step === 1 ? (
          <form onSubmit={handleRegisterSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="username" className="text-sm font-medium text-slate-200">
                Username
              </label>
              <input
                id="username"
                name="username"
                autoComplete="username"
                className={inputClass}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium text-slate-200">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                className={inputClass}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="country" className="text-sm font-medium text-slate-200">
                Country
              </label>
              <select
                id="country"
                name="country"
                className={inputClass}
                value={country}
                onChange={(e) => setCountry(e.target.value)}
              >
                {COUNTRIES.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <label htmlFor="phone" className="text-sm font-medium text-slate-200">
                Phone number
              </label>
              <input
                id="phone"
                name="phone"
                type="tel"
                autoComplete="tel"
                placeholder="e.g. 09171234567 or 639171234567"
                className={inputClass}
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
                required
              />
              <p className="text-xs text-slate-500">
                Digits only: use your local number with a leading 0 (e.g. PH: 09…) or full
                international form (63…). Pick country first, then auto-fill for a sample
                national number.
              </p>
            </div>
            <Button
              type="button"
              variant="secondary"
              className="w-full"
              onClick={fillRandomCredentials}
              disabled={loading}
            >
              Auto-fill username, password &amp; phone
            </Button>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Sending code…" : "Register & send code"}
            </Button>
          </form>
        ) : null}

        {step === 2 ? (
          <form onSubmit={handleOtpSubmit} className="space-y-4">
            <p className="text-sm text-slate-300">
              Enter the 6-digit code sent to{" "}
              <span className="font-medium text-white">{phoneNumber}</span> for{" "}
              <span className="font-medium text-white">{username}</span>.
            </p>
            <div className="space-y-2">
              <label htmlFor="otp" className="text-sm font-medium text-slate-200">
                Verification code
              </label>
              <input
                id="otp"
                name="otp"
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="000000"
                className={`${inputClass} text-center text-lg tracking-[0.4em] font-mono`}
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                required
              />
            </div>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={resetFlow}
                disabled={loading}
              >
                Start over
              </Button>
              <Button type="submit" className="flex-1" disabled={loading || otp.length < 6}>
                {loading ? "Verifying…" : "Verify"}
              </Button>
            </div>
          </form>
        ) : null}

        {step === 3 ? (
          <div className="space-y-4">
            <div className="rounded-lg border border-emerald-900/50 bg-emerald-950/30 px-3 py-3 text-sm text-emerald-100">
              <p className="font-medium text-emerald-50 mb-2">Account verified</p>
              <ul className="space-y-1 text-emerald-100/90">
                <li>
                  <span className="text-emerald-200/70">Username</span> {username}
                </li>
                <li>
                  <span className="text-emerald-200/70">Phone</span>{" "}
                  {registeredSummary?.phone_number ?? phoneNumber}
                </li>
                <li>
                  <span className="text-emerald-200/70">Country</span> {country}
                </li>
              </ul>
            </div>
            <Button type="button" variant="outline" className="w-full" onClick={resetFlow}>
              Register another account
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  )
}
