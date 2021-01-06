CREATE TABLE users (
    username text PRIMARY KEY,
    billing_id text,
    user_group_id uuid,
    status int,
    password text,
    ips frozen<list<text>>
);

CREATE TABLE users_by_ip (
    ip text, 
    username text,
    PRIMARY KEY (ip, username)
);


INSERT INTO users (username, password, billing_id, user_group_id) VALUES('josh', 'password123', '', 72e2d683-3e94-4cab-a266-04abd30ce1bd);