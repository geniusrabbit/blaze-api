
CREATE OR REPLACE FUNCTION updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now(); 
   RETURN NEW;
END;
$$ language 'plpgsql';

-- @link https://github.com/geniusrabbit/notificationcenter/tree/master/pg
CREATE OR REPLACE FUNCTION notify_update_event() RETURNS TRIGGER AS $$
    DECLARE
        data json;
        notification json;
    BEGIN

        -- Convert the old or new row to JSON, based on the kind of action.
        -- Action = DELETE?             -> OLD row
        -- Action = INSERT or UPDATE?   -> NEW row
        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;

        -- Contruct the notification as a JSON string.
        notification = json_build_object(
                          'schema', TG_TABLE_SCHEMA,
                          'table',  TG_TABLE_NAME,
                          'action', TG_OP,
                          'data',   data);


        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('update_events', notification::text);

        -- Result is ignored since this is an AFTER trigger
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
