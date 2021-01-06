CREATE TABLE user_groups (
    id uuid PRIMARY KEY,
    name text,
    domain_blacklist frozen<map<text, list<text>>>,
    ip_blacklist list<text>
);

INSERT INTO user_groups (id, name, domain_blacklist) VALUES(uuid(), 'default', { 'www.google.com': ['216.58.194.196'] });