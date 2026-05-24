-- Buat sequence untuk QIOS-ID supaya atomic, tidak ada race condition.
-- Advance ke nilai MAX yang sudah ada supaya tidak konflik dengan data lama.

CREATE SEQUENCE IF NOT EXISTS qios_id_seq
    START 1
    INCREMENT 1
    NO MAXVALUE
    NO CYCLE;

SELECT setval(
    'qios_id_seq',
    COALESCE(
        MAX(CAST(SUBSTRING(qios_id FROM 6) AS INTEGER)),
        0
    )
)
FROM businesses
WHERE qios_id ~ '^QIOS-\d+$';
