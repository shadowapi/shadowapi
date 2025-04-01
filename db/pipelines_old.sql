
CREATE TABLE pipeline(
                         "uuid"            UUID PRIMARY KEY,
                         user_uuid    UUID,
                         name              VARCHAR NOT NULL,
                         flow              JSONB NOT NULL,

                         created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                         updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE pipeline_entry(
                               "uuid"            UUID PRIMARY KEY,
                               pipeline_uuid     UUID NOT NULL,
                               parent_uuid       UUID,
                               "type"            TEXT NOT NULL,
                               params            JSONB,

                               created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               updated_at TIMESTAMP WITH TIME ZONE,

                               CONSTRAINT fk_pipeline_entry_pipeline_uuid
                                   FOREIGN KEY(pipeline_uuid) REFERENCES pipeline("uuid") ON DELETE CASCADE
);