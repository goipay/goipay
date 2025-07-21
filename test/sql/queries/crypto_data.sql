-- name: FindJoinedCryptoDataByUserId :one
SELECT * FROM crypto_data as cd
JOIN xmr_crypto_data as xcd ON cd.xmr_id = xcd.id
WHERE user_id = $1;

-- name: UpdateIndicesXMRCryptoDataById :one
UPDATE xmr_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;


-- name: UpdateIndicesBTCCryptoDataById :one
UPDATE btc_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;

-- name: UpdateIndicesLTCCryptoDataById :one
UPDATE ltc_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;

-- name: UpdateIndicesETHCryptoDataById :one
UPDATE eth_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;