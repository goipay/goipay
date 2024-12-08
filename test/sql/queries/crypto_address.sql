-- name: FindCryptoAddressByAddress :one
SELECT * FROM crypto_addresses
WHERE address = $1;