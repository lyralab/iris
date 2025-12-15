ALTER TABLE alerts
    ALTER COLUMN method TYPE TEXT using method::TEXT;