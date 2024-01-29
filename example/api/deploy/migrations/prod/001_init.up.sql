
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

-- @link https://schinckel.net/2021/09/09/automatically-expire-rows-in-postgres/
CREATE OR REPLACE FUNCTION keep_for() RETURNS TRIGGER AS $$
    -- param[0] = column name
    -- param[1] = interval
    BEGIN
        IF TG_WHEN <> 'BEFORE' THEN
            RAISE EXCEPTION 'keep_for() may only run as a BEFORE trigger';
        END IF ;

        IF TG_ARGV[0]::TEXT IS NULL THEN
            RAISE EXCEPTION 'keep_for() must be installed with a column name to check';
        END IF;

        IF TG_ARGV[1]::INTERVAL IS NULL THEN
            RAISE EXCEPTION 'keep_for() must be installed with an INTERVAL to keep data for';
        END IF;

        IF TG_OP = 'INSERT' THEN

            EXECUTE 'DELETE FROM ' || quote_ident(TG_TABLE_NAME) ||
                ' WHERE ' || TG_ARGV[0]::TEXT ||
                    ' < now() - INTERVAL ' || quote_literal(TG_ARGV[1]::TEXT) || ';';

        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;
