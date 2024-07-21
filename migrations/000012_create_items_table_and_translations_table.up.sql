CREATE TABLE items (
                       id BIGSERIAL PRIMARY KEY,
                       category_id BIGINT DEFAULT 100 REFERENCES categories(id) ON DELETE SET DEFAULT,
                       created_at TIMESTAMP(0) WITH TIME ZONE DEFAULT NOW() NOT NULL
);
create table item_translations
(
    id           serial primary key,
    item_id  bigint references items(id) on delete cascade,
    language_id  integer references languages(id) on delete cascade,
    translation  text not null,
    image text not null default  'https://placehold.co/600x400',
    unique (item_id, language_id)
);