-- Table: xchat_message

DROP TABLE IF EXISTS xchat_message ;

CREATE TABLE xchat_message
(
  chat_id bigint NOT NULL,
  chat_type character varying(10) NOT NULL,
  id bigint NOT NULL,
  uid character varying(32) NOT NULL,
  ts timestamp with time zone NOT NULL,
  msg text NOT NULL,
  domain character varying(16) NOT NULL,
  CONSTRAINT xchat_message_pkey PRIMARY KEY (chat_id, chat_type, id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE xchat_message
  OWNER TO xchat;


CREATE INDEX xchat_message_chat_id ON xchat_message USING btree(chat_id, chat_type);
CREATE INDEX xchat_message_ts ON xchat_message USING btree(ts);
CREATE INDEX xchat_message_id ON xchat_message USING btree(id);
