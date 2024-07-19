alter table category_translations drop column IF EXISTS  image;
alter table categories add column image Text NOT NULL default 'https://placehold.co/600x400' ;
