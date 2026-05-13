"""
z-uc Sub-Secret Generator & Verifier

Usage:
    python subsecret_generator.py <appId>
    python subsecret_generator.py <appId> --decrypt <cipher_hex> <plain_union_id>

Principle:
    master_secret  =  JwtSecret (hardcoded, same as Go models/token.go)
    app_key        =  HMAC-SHA256(master_secret, "z-uc:app:<appId>")  → 32 bytes
    cipher         =  AES-256-GCM(key=app_key, plaintext=union_id)
                      → hex(nonce[12] + ciphertext + tag[16])
    verify:  decrypt(cipher) == plain_union_id

Go counterpart:  models/token.go → DeriveAppSecret / EncryptAppPayload / DecryptAppPayload
"""

import sys
import os
import hmac
import hashlib
import secrets
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

JWT_SECRET = os.environ.get("JWT_SECRET", "REAAW332LPPPPPEC00S++++++SDEDSSDFCCCCCC_____FFRFDSSDS").encode()
PREFIX     = b"z-uc:app:"


def derive_app_key(app_id: str) -> bytes:
    return hmac.new(JWT_SECRET, PREFIX + app_id.encode(), hashlib.sha256).digest()


def encrypt_union_id(app_id: str, union_id: str) -> str:
    key = derive_app_key(app_id)
    nonce = secrets.token_bytes(12)
    aesgcm = AESGCM(key)
    ciphertext = aesgcm.encrypt(nonce, union_id.encode(), None)
    return (nonce + ciphertext).hex()


def decrypt_union_id(app_id: str, cipher_hex: str) -> str:
    key = derive_app_key(app_id)
    raw = bytes.fromhex(cipher_hex)
    nonce, ciphertext = raw[:12], raw[12:]
    aesgcm = AESGCM(key)
    plaintext = aesgcm.decrypt(nonce, ciphertext, None)
    return plaintext.decode()


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python subsecret_generator.py <appId>")
        print("  python subsecret_generator.py <appId> --decrypt <cipher_hex> <plain_union_id>")
        sys.exit(1)

    app_id = sys.argv[1]

    if len(sys.argv) >= 5 and sys.argv[2] == "--decrypt":
        cipher_hex = sys.argv[3]
        plain_union_id = sys.argv[4]
        decrypted = decrypt_union_id(app_id, cipher_hex)
        match = decrypted == plain_union_id
        print(f"appId:     {app_id}")
        print(f"decrypted: {decrypted}")
        print(f"expected:  {plain_union_id}")
        print(f"match:     {match}")
        sys.exit(0 if match else 1)

    if len(sys.argv) >= 4 and sys.argv[2] == "--encrypt":
        union_id = sys.argv[3]
        cipher = encrypt_union_id(app_id, union_id)
        print(f"appId:    {app_id}")
        print(f"union_id: {union_id}")
        print(f"cipher:   {cipher}")
        sys.exit(0)

    app_key = derive_app_key(app_id)
    print(f"appId:    {app_id}")
    print(f"app_key:  {app_key.hex()}")
    print(f"  Go equivalent: models.DeriveAppSecret(\"{app_id}\")")
