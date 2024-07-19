alter table categories drop column image;
alter table category_translations add column image Text NOT NULL default 'https://placehold.co/600x400' ;
