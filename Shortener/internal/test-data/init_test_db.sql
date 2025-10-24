
DROP TABLE IF EXISTS ANALYTIC;
DROP TABLE IF EXISTS URL;

CREATE TABLE IF NOT EXISTS URL (
    ID SERIAL PRIMARY KEY,
    URL VARCHAR(255) NOT NULL,
    ALIAS VARCHAR(128) UNIQUE
);

CREATE TABLE IF NOT EXISTS ANALYTIC (
    ID SERIAL PRIMARY KEY,
    ALIAS VARCHAR(128),
    USER_AGENT VARCHAR(255),
    REQUEST_TIME TIMESTAMP,
    CONSTRAINT ALIAS_NAME FOREIGN KEY (ALIAS) REFERENCES URL (ALIAS)
);


INSERT INTO URL (URL, ALIAS) VALUES
('https://google.com', 'ggl'),
('https://github.com', 'gh'),
('https://openai.com', 'ai'),
('https://youtube.com', 'yt'),
('https://news.ycombinator.com', 'hn'),
('https://example.com', 'ex'),
('https://postgresql.org', 'pg'),
('https://abc.com', 'abc'),
('https://golang.org', 'go');


INSERT INTO ANALYTIC (ALIAS, USER_AGENT, REQUEST_TIME) VALUES
('ggl', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', NOW() - INTERVAL '1 day'),
('ggl', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', NOW() - INTERVAL '2 hours'),
('gh', 'Mozilla/5.0 (X11; Linux x86_64)', NOW() - INTERVAL '5 hours'),
('gh', 'curl/8.0.1', NOW() - INTERVAL '1 hour'),
('ai', 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)', NOW() - INTERVAL '30 minutes'),
('yt', 'Mozilla/5.0 (Android 14; Mobile; rv:109.0)', NOW() - INTERVAL '10 minutes'),
('yt', 'Mozilla/5.0 (SmartTV; Linux; Tizen 7.0)', NOW() - INTERVAL '2 days'),
('hn', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', NOW() - INTERVAL '12 hours'),
('ex', 'PostmanRuntime/7.33.0', NOW() - INTERVAL '45 minutes'),
('pg', 'psql-client/16.0', NOW() - INTERVAL '3 days'),
('go', 'Go-http-client/2.0', NOW() - INTERVAL '5 minutes'),
('abc', 'test-ua', NOW() - INTERVAL '15 minutes');