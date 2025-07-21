-- name: CreateCryptoData :one
INSERT INTO crypto_data(xmr_id, btc_id, ltc_id, eth_id, bnb_id, user_id) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: FindCryptoDataByUserId :one
SELECT * FROM crypto_data 
WHERE user_id = $1;

-- name: SetXMRCryptoDataByUserId :one
UPDATE crypto_data
SET xmr_id = $2 
WHERE user_id = $1
RETURNING *;

-- name: SetBTCCryptoDataByUserId :one
UPDATE crypto_data
SET btc_id = $2
WHERE user_id = $1
RETURNING *;

-- name: SetLTCCryptoDataByUserId :one
UPDATE crypto_data
SET ltc_id = $2
WHERE user_id = $1
RETURNING *;

-- name: SetETHCryptoDataByUserId :one
UPDATE crypto_data
SET eth_id = $2
WHERE user_id = $1
RETURNING *;

-- name: SetBNBCryptoDataByUserId :one
UPDATE crypto_data
SET bnb_id = $2
WHERE user_id = $1
RETURNING *;


-- XMR
-- name: CreateXMRCryptoData :one
INSERT INTO xmr_crypto_data(priv_view_key, pub_spend_key) VALUES ($1, $2)
RETURNING *;

-- name: FindKeysXMRCryptoDataById :one
SELECT priv_view_key, pub_spend_key
FROM xmr_crypto_data
WHERE id = $1;

-- name: UpdateKeysXMRCryptoDataById :one
UPDATE xmr_crypto_data
SET priv_view_key = $2,
    pub_spend_key = $3,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindKeysAndIncrementedIndicesXMRCryptoDataById :one
UPDATE xmr_crypto_data
SET last_minor_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN 0
        ELSE last_minor_index + 1
    END,
    last_major_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN last_major_index + 1
        ELSE last_major_index
    END
WHERE id = $1
RETURNING priv_view_key, pub_spend_key, last_major_index, last_minor_index;


-- BTC
-- name: CreateBTCCryptoData :one
INSERT INTO btc_crypto_data(master_pub_key) VALUES ($1)
RETURNING *;

-- name: UpdateKeysBTCCryptoDataById :one
UPDATE btc_crypto_data
SET master_pub_key = $2,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindKeysAndIncrementedIndicesBTCCryptoDataById :one
UPDATE btc_crypto_data
SET last_minor_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN 0
        ELSE last_minor_index + 1
    END,
    last_major_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN last_major_index + 1
        ELSE last_major_index
    END
WHERE id = $1
RETURNING master_pub_key, last_major_index, last_minor_index;


-- LTC
-- name: CreateLTCCryptoData :one
INSERT INTO ltc_crypto_data(master_pub_key) VALUES ($1)
RETURNING *;

-- name: UpdateKeysLTCCryptoDataById :one
UPDATE ltc_crypto_data
SET master_pub_key = $2,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindKeysAndIncrementedIndicesLTCCryptoDataById :one
UPDATE ltc_crypto_data
SET last_minor_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN 0
        ELSE last_minor_index + 1
    END,
    last_major_index = CASE 
        WHEN last_minor_index >= 2147483647 THEN last_major_index + 1
        ELSE last_major_index
    END
WHERE id = $1
RETURNING master_pub_key, last_major_index, last_minor_index;


-- ETH
-- name: CreateETHCryptoData :one
INSERT INTO eth_crypto_data(master_pub_key) VALUES ($1)
RETURNING *;

-- name: FindKeysAndLockETHCryptoDataById :one
SELECT master_pub_key
FROM eth_crypto_data
WHERE id = $1
FOR SHARE;

-- name: UpdateKeysETHCryptoDataById :one
UPDATE eth_crypto_data
SET master_pub_key = $2,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindIndicesAndLockETHCryptoDataById :one
SELECT last_major_index, last_minor_index 
FROM eth_crypto_data
WHERE id = $1
FOR UPDATE;

-- name: UpdateIndicesETHCryptoDataById :one
UPDATE eth_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;

-- BNB
-- name: CreateBNBCryptoData :one
INSERT INTO bnb_crypto_data(master_pub_key) VALUES ($1)
RETURNING *;

-- name: FindKeysAndLockBNBCryptoDataById :one
SELECT master_pub_key
FROM bnb_crypto_data
WHERE id = $1
FOR SHARE;

-- name: UpdateKeysBNBCryptoDataById :one
UPDATE bnb_crypto_data
SET master_pub_key = $2,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindIndicesAndLockBNBCryptoDataById :one
SELECT last_major_index, last_minor_index 
FROM bnb_crypto_data
WHERE id = $1
FOR UPDATE;

-- name: UpdateIndicesBNBCryptoDataById :one
UPDATE bnb_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;