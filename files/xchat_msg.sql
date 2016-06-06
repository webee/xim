-- Table: xchat_message

DROP TABLE IF EXISTS xchat_message ;

CREATE TABLE xchat_message
(
  chat_id bigint NOT NULL,
  id bigint NOT NULL,
  uid character varying(32) NOT NULL,
  ts timestamp with time zone NOT NULL,
  msg text NOT NULL,
  CONSTRAINT xchat_message_pkey PRIMARY KEY (chat_id, id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE xchat_message
  OWNER TO xchat;


CREATE INDEX xchat_message_ts ON xchat_message USING btree(ts);
