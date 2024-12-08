-- name: FindAllInvoices :many
SELECT * FROM invoices;

-- name: FindAllInvoicesByIds :many
SELECT * FROM invoices
WHERE id = ANY($1::uuid[]);

-- name: FindInvoiceById :one
SELECT * FROM invoices
WHERE id = $1;