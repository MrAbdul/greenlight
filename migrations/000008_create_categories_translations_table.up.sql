alter table categories drop column title;
create table languages
(
    id   serial primary key,
    code text    not null, -- e.g., 'en', 'fr', 'es'
    name text    not null  -- e.g., 'English', 'French', 'Spanish'
);

create table category_translations
(
    id           serial primary key,
    category_id  bigint references categories(id) on delete cascade,
    language_id  integer references languages(id) on delete cascade,
    translation  text not null,
    unique (category_id, language_id)
);


insert into languages (id, code, name) values
                                        (1, 'en', 'English'),
                                        (2, 'ar', 'Arabic');

