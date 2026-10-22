#!/usr/bin/env node
/**
 * Register a Vaultwarden / Bitwarden user via the public register API,
 * then mint a Personal API Key for smoke (API backend).
 * Crypto matches Bitwarden clients (PBKDF2 + AES-CBC-256-HMAC-SHA256 + RSA-2048).
 *
 * Usage:
 *   node scripts/smoke-register.mjs <serverUrl> <email> <password> [name]
 *
 * On success prints (stdout):
 *   BWSF_API_CLIENT_ID=user.<uuid>
 *   BWSF_API_CLIENT_SECRET=<apiKey>
 *
 * Exit 0 if created or already exists; non-zero on hard failures.
 */
import crypto from "node:crypto";
import https from "node:https";
import http from "node:http";
import { URL } from "node:url";

const KDF_TYPE_PBKDF2 = 0;
const KDF_ITERATIONS = 600000;

function pbkdf2(password, salt, iterations, keyLen) {
  return crypto.pbkdf2Sync(
    Buffer.from(password, "utf8"),
    Buffer.from(salt, "utf8"),
    iterations,
    keyLen,
    "sha256",
  );
}

function hkdfExpand(prk, info, length) {
  const infoBuf = Buffer.from(info, "utf8");
  let t = Buffer.alloc(0);
  let okm = Buffer.alloc(0);
  let i = 0;
  while (okm.length < length) {
    i += 1;
    const hmac = crypto.createHmac("sha256", prk);
    hmac.update(t);
    hmac.update(infoBuf);
    hmac.update(Buffer.from([i]));
    t = hmac.digest();
    okm = Buffer.concat([okm, t]);
  }
  return okm.subarray(0, length);
}

function stretchMasterKey(masterKey) {
  const encKey = hkdfExpand(masterKey, "enc", 32);
  const macKey = hkdfExpand(masterKey, "mac", 32);
  return { encKey, macKey };
}

function encryptAes256Hmac(plaintext, encKey, macKey) {
  const iv = crypto.randomBytes(16);
  const cipher = crypto.createCipheriv("aes-256-cbc", encKey, iv);
  const ct = Buffer.concat([cipher.update(plaintext), cipher.final()]);
  const mac = crypto.createHmac("sha256", macKey).update(iv).update(ct).digest();
  return `2.${iv.toString("base64")}|${ct.toString("base64")}|${mac.toString("base64")}`;
}

function makeMasterPasswordHash(email, password) {
  const normalizedEmail = email.trim().toLowerCase();
  const masterKey = pbkdf2(password, normalizedEmail, KDF_ITERATIONS, 32);
  return {
    normalizedEmail,
    masterKey,
    masterPasswordHash: pbkdf2(masterKey, password, 1, 32).toString("base64"),
  };
}

function makeRegistrationPayload(email, password, name) {
  const { normalizedEmail, masterKey, masterPasswordHash } =
    makeMasterPasswordHash(email, password);
  const { encKey, macKey } = stretchMasterKey(masterKey);

  const symKey = crypto.randomBytes(64);
  const protectedSymKey = encryptAes256Hmac(symKey, encKey, macKey);

  const { publicKey, privateKey } = crypto.generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicExponent: 0x10001,
  });
  const publicKeySpki = publicKey.export({ type: "spki", format: "der" }).toString("base64");
  const privateKeyPkcs8 = privateKey.export({ type: "pkcs8", format: "der" });

  const symEncKey = symKey.subarray(0, 32);
  const symMacKey = symKey.subarray(32, 64);
  const protectedPrivateKey = encryptAes256Hmac(privateKeyPkcs8, symEncKey, symMacKey);

  return {
    name: name || "Smoke",
    email: normalizedEmail,
    masterPasswordHash,
    masterPasswordHint: null,
    key: protectedSymKey,
    kdf: KDF_TYPE_PBKDF2,
    kdfIterations: KDF_ITERATIONS,
    keys: {
      publicKey: publicKeySpki,
      encryptedPrivateKey: protectedPrivateKey,
    },
  };
}

function request(method, urlString, { body, headers } = {}) {
  const url = new URL(urlString);
  const lib = url.protocol === "https:" ? https : http;
  const payload =
    body == null
      ? null
      : typeof body === "string"
        ? body
        : JSON.stringify(body);

  const opts = {
    method,
    hostname: url.hostname,
    port: url.port || (url.protocol === "https:" ? 443 : 80),
    path: `${url.pathname}${url.search}`,
    headers: {
      Accept: "application/json",
      ...(headers || {}),
      ...(payload
        ? {
            "Content-Type":
              headers?.["Content-Type"] || "application/json",
            "Content-Length": Buffer.byteLength(payload),
          }
        : {}),
    },
    rejectUnauthorized: false, // smoke self-signed; CA is for bw/bwsf path
  };

  return new Promise((resolve, reject) => {
    const req = lib.request(opts, (res) => {
      const chunks = [];
      res.on("data", (c) => chunks.push(c));
      res.on("end", () => {
        resolve({
          status: res.statusCode || 0,
          body: Buffer.concat(chunks).toString("utf8"),
          headers: res.headers,
        });
      });
    });
    req.on("error", reject);
    if (payload) req.write(payload);
    req.end();
  });
}

function requestJson(method, urlString, body, headers) {
  return request(method, urlString, { body, headers });
}

function isAlreadyExists(status, body) {
  if (status !== 400 && status !== 409) return false;
  const lower = body.toLowerCase();
  return (
    lower.includes("already") ||
    lower.includes("exists") ||
    lower.includes("taken") ||
    lower.includes("duplicate")
  );
}

async function registerAccount(base, email, password, name) {
  const payload = makeRegistrationPayload(email, password, name);
  // Vaultwarden 1.37.4+ removed legacy /identity/accounts/register;
  // current clients use register/finish.
  const endpoints = [
    `${base}/identity/accounts/register/finish`,
    `${base}/identity/accounts/register`,
    `${base}/api/accounts/register`,
  ];

  let last = null;
  for (const endpoint of endpoints) {
    last = await requestJson("POST", endpoint, payload);
    if (last.status >= 200 && last.status < 300) {
      console.error(`Registered ${email} via ${endpoint}`);
      return;
    }
    if (isAlreadyExists(last.status, last.body)) {
      console.error(`User already exists: ${email}`);
      return;
    }
  }

  throw new Error(
    `Register failed (last status ${last?.status}): ${last?.body || "no response"}`,
  );
}

function decodeJwtSub(accessToken) {
  const parts = accessToken.split(".");
  if (parts.length < 2) return null;
  const json = Buffer.from(parts[1], "base64url").toString("utf8");
  try {
    const payload = JSON.parse(json);
    return payload.sub || payload.Name || null;
  } catch {
    return null;
  }
}

async function passwordLogin(base, email, masterPasswordHash) {
  const form = new URLSearchParams({
    grant_type: "password",
    username: email.trim().toLowerCase(),
    password: masterPasswordHash,
    scope: "api offline_access",
    client_id: "web",
    deviceType: "8",
    deviceIdentifier: crypto.randomUUID(),
    deviceName: "bwsf-smoke",
  });
  const res = await request("POST", `${base}/identity/connect/token`, {
    body: form.toString(),
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
  });
  if (res.status < 200 || res.status >= 300) {
    throw new Error(
      `Password login failed (HTTP ${res.status}): ${res.body}`,
    );
  }
  const parsed = JSON.parse(res.body);
  if (!parsed.access_token) {
    throw new Error("Password login missing access_token");
  }
  return parsed.access_token;
}

async function fetchApiKey(base, accessToken, masterPasswordHash) {
  const res = await requestJson(
    "POST",
    `${base}/api/accounts/api-key`,
    { masterPasswordHash },
    { Authorization: `Bearer ${accessToken}` },
  );
  if (res.status < 200 || res.status >= 300) {
    throw new Error(
      `Create API key failed (HTTP ${res.status}): ${res.body}`,
    );
  }
  const parsed = JSON.parse(res.body);
  if (!parsed.apiKey) {
    throw new Error("API key response missing apiKey");
  }
  return parsed.apiKey;
}

async function resolveUserId(base, accessToken, jwtSub) {
  if (jwtSub && String(jwtSub).length > 0) {
    return String(jwtSub);
  }
  const res = await request("GET", `${base}/api/accounts/profile`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  if (res.status < 200 || res.status >= 300) {
    throw new Error(
      `Profile fetch failed (HTTP ${res.status}): ${res.body}`,
    );
  }
  const parsed = JSON.parse(res.body);
  const id = parsed.id || parsed.Id;
  if (!id) {
    throw new Error("Profile response missing id");
  }
  return String(id);
}

async function bootstrap(serverUrl, email, password, name) {
  const base = serverUrl.replace(/\/+$/, "");
  await registerAccount(base, email, password, name);

  const { masterPasswordHash } = makeMasterPasswordHash(email, password);
  const accessToken = await passwordLogin(base, email, masterPasswordHash);
  const userId = await resolveUserId(
    base,
    accessToken,
    decodeJwtSub(accessToken),
  );
  const apiKey = await fetchApiKey(base, accessToken, masterPasswordHash);

  const clientId = userId.startsWith("user.") ? userId : `user.${userId}`;
  // Machine-readable for the smoke shell (stdout). Logs go to stderr.
  console.log(`BWSF_API_CLIENT_ID=${clientId}`);
  console.log(`BWSF_API_CLIENT_SECRET=${apiKey}`);
}

const [serverUrl, email, password, name] = process.argv.slice(2);
if (!serverUrl || !email || !password) {
  console.error(
    "Usage: node scripts/smoke-register.mjs <serverUrl> <email> <password> [name]",
  );
  process.exit(2);
}

bootstrap(serverUrl, email, password, name).catch((err) => {
  console.error(err.message || err);
  process.exit(1);
});
