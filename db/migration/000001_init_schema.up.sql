CREATE TABLE "links" (
    "short_url" VARCHAR PRIMARY KEY,
    "long_url" VARCHAR UNIQUE NOT NULL
);