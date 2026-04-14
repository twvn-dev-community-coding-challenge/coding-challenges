export const generateSixDigitOtp = (): string => {
  const buf = new Uint32Array(1);
  crypto.getRandomValues(buf);
  const n = (buf[0] ?? 0) % 1_000_000;
  return n.toString().padStart(6, '0');
};
