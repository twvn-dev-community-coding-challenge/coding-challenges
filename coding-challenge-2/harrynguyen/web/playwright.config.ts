import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "@playwright/test";

const __dirname = dirname(fileURLToPath(import.meta.url));

/** Load repo-root `.env` so local e2e pick up `SMS_INTERNAL_API_KEY` (same as the API server). Does not override existing env. */
function loadEnvFile(filePath: string) {
  if (!existsSync(filePath)) return;
  const text = readFileSync(filePath, "utf8");
  for (const line of text.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eq = trimmed.indexOf("=");
    if (eq <= 0) continue;
    const key = trimmed.slice(0, eq).trim();
    let val = trimmed.slice(eq + 1).trim();
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    if (process.env[key] === undefined) {
      process.env[key] = val;
    }
  }
}

loadEnvFile(resolve(__dirname, "../.env"));

const apiBaseUrl = process.env.PLAYWRIGHT_API_BASE_URL || "http://127.0.0.1:8080";

const smsInternalKey =
  process.env.PW_SMS_INTERNAL_API_KEY || process.env.SMS_INTERNAL_API_KEY;

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  fullyParallel: false,
  retries: process.env.CI ? 2 : 0,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: apiBaseUrl,
    trace: "on-first-retry",
    ...(smsInternalKey
      ? { extraHTTPHeaders: { Authorization: `Bearer ${smsInternalKey}` } }
      : {}),
  },
});
