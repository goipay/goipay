-- name: FindJoinedCryptoDataByUserId :one
SELECT * FROM crypto_data as cd
JOIN xmr_crypto_data as xcd ON cd.xmr_id = xcd.id
WHERE user_id = $1;