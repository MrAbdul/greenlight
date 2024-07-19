alter table categories add column title text NOT NULL default 'default' ;
drop table if exists category_translations;
drop table if exists languages;
