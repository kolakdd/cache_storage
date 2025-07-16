CREATE TABLE "User" (
    id UUID PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    hash_password VARCHAR(255) NOT NULL,
    created_at timestamp without time zone NOT NULL default now(),
    updated_at timestamp without time zone
);

CREATE TABLE "Object" (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES "User" (id) NOT NULL,
    name VARCHAR(255) NOT NULL,
    mimetype VARCHAR(100),
    public BOOLEAN DEFAULT FALSE,
    size BIGINT NOT NULL,
    upload_s3 BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE, -- soft delete
    eliminated BOOLEAN DEFAULT FALSE, -- hard delete from s3
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone
);

CREATE TABLE "UserXObject" (
    user_id UUID NOT NULL REFERENCES "User" (id),
    object_id UUID NOT NULL REFERENCES "Object" (id) ON DELETE CASCADE,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone,
    PRIMARY KEY (user_id, object_id)
);