// src/wasm.d.ts
// ────────────────────────────────────────────────────────────────────────────────
// Type declarations for Go → JavaScript exports from umbra.wasm
// These appear on globalThis / window after successful go.run()
//
// Usage example:
//   if (window.umbraReady?.()) { ... }
//   const result = window.MACE_Encrypt?.(keyUint8, dataUint8, "my-context", 12);
// ────────────────────────────────────────────────────────────────────────────────

/**
 * Represents the result object returned by MACE_Encrypt
 */
interface MACE_EncryptResult {
  /**
   * Encrypted ciphertext as Uint8Array
   */
  cipher: Uint8Array;

  /**
   * Salt used in the encryption (for decryption/verification later)
   */
  salt: Uint8Array;
}

/**
 * Umbra WASM module exports (exposed on globalThis / window)
 */
interface UmbraExports {
  /**
   * Simple readiness check — returns a string confirming initialization.
   * Useful for confirming WASM has loaded and run successfully.
   *
   * @returns "Umbra WASM initialized"
   */
  umbraReady(): string;

  /**
   * Encrypts data using MACE (with PBKDF2-derived key stretching).
   *
   * @param key - The raw encryption key (must be appropriate length for the cipher,
   *              typically 32 bytes for AES-256). Passed as Uint8Array.
   * @param data - The plaintext data to encrypt. Uint8Array recommended.
   * @param context - A string context used in key derivation (e.g. "app:login" or domain name).
   * @param difficulty - PBKDF2 iteration count multiplier (higher = slower but more secure).
   *                     Usually 8–20 range for client-side use.
   *
   * @returns Object containing { cipher: Uint8Array, salt: Uint8Array }
   * @throws JS Error (via console.error) if:
   *   - Fewer than 4 arguments
   *   - key or data is not a valid TypedArray / cannot be copied
   */
  MACE_Encrypt(
    key: Uint8Array,
    data: Uint8Array,
    context: string,
    difficulty: number
  ): MACE_EncryptResult;
}

const umbraReady: UmbraExports['umbraReady'] | undefined;
const MACE_Encrypt: UmbraExports['MACE_Encrypt'] | undefined;
