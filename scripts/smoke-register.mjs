#!/usr/bin/env node
/**
 * Register a Vaultwarden / Bitwarden user via the public register API.
 * Crypto matches Bitwarden clients (PBKDF2 + AES-CBC-256-HMAC-SHA256 + RSA-2048).
 *
 * Usage:
 *   node scripts/smoke-register.mjs <serverUrl> <email> <password> [name]
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

function makeRegistrationPayload(email, password, name) {
  const normalizedEmail = email.trim().toLowerCase();
  const masterKey = pbkdf2(password, normalizedEmail, KDF_ITERATIONS, 32);
  const masterPasswordHash = pbkdf2(masterKey, password, 1, 32).toString("base64");
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

function requestJson(method, urlString, body) {
  const url = new URL(urlString);
  const lib = url.protocol === "https:" ? https : http;
  const payload = body == null ? null : JSON.stringify(body);

  const opts = {
    method,
    hostname: url.hostname,
    port: url.port || (url.protocol === "https:" ? 443 : 80),
    path: `${url.pathname}${url.search}`,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(payload ? { "Content-Length": Buffer.byteLength(payload) } : {}),
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
        });
      });
    });
    req.on("error", reject);
    if (payload) req.write(payload);
    req.end();
  });
}

async function register(serverUrl, email, password, name) {
  const base = serverUrl.replace(/\/+$/, "");
  const payload = makeRegistrationPayload(email, password, name);
  const endpoints = [
    `${base}/api/accounts/register`,
    `${base}/identity/accounts/register`,
  ];

  let last = null;
  for (const endpoint of endpoints) {
    last = await requestJson("POST", endpoint, payload);
    if (last.status >= 200 && last.status < 300) {
      console.log(`Registered ${email} via ${endpoint}`);
      return;
    }
    // Already exists / conflict → treat as success for idempotent bootstrap
    if (last.status === 400 || last.status === 409) {
      const lower = last.body.toLowerCase();
      if (
        lower.includes("already") ||
        lower.includes("exists") ||
        lower.includes("taken") ||
        lower.includes("duplicate")
      ) {
        console.log(`User already exists: ${email}`);
        return;
      }
    }
  }

  throw new Error(
    `Register failed (last status ${last?.status}): ${last?.body || "no response"}`,
  );
}

const [serverUrl, email, password, name] = process.argv.slice(2);
if (!serverUrl || !email || !password) {
  console.error(
    "Usage: node scripts/smoke-register.mjs <serverUrl> <email> <password> [name]",
  );
  process.exit(2);
}

register(serverUrl, email, password, name).catch((err) => {
  console.error(err.message || err);
  process.exit(1);
});
