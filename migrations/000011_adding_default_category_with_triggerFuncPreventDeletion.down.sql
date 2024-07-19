-- Drop the trigger to prevent default category deletion
DROP TRIGGER IF EXISTS prevent_default_category_deletion_trigger ON categories;

-- Drop the function to prevent default category deletion
DROP FUNCTION IF EXISTS prevent_default_category_deletion;

-- Delete the default category translations
DELETE FROM category_translations WHERE category_id = 100;

-- Delete the default category
DELETE FROM categories WHERE id = 100;

-- Update the sequence to continue from the highest existing ID
SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT COALESCE(MAX(id), 1) FROM categories) + 1);