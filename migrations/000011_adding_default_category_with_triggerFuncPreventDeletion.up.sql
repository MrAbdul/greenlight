-- Insert the default category with a specific ID
INSERT INTO categories (id, version) VALUES (100, 1);

-- Insert translations for the default category
INSERT INTO category_translations (category_id, language_id, translation, image)
VALUES (100, (SELECT id FROM languages WHERE code = 'en'), 'Uncategorized', 'https://placehold.co/600x400'),
       (100, (SELECT id FROM languages WHERE code = 'ar'), 'غير مصنف', 'https://placehold.co/600x400');

-- Update the sequence to ensure it does not conflict with the manually set ID
SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT MAX(id) FROM categories) + 1);

-- Create a function to prevent the deletion of the default category
CREATE OR REPLACE FUNCTION prevent_default_category_deletion()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD.id = 100 THEN
        RAISE EXCEPTION 'Cannot delete the default category';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to call the function before deleting a category
CREATE TRIGGER prevent_default_category_deletion_trigger
    BEFORE DELETE ON categories
    FOR EACH ROW
EXECUTE FUNCTION prevent_default_category_deletion();